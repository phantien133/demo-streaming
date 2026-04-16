package streamkeys

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authpkg "demo-streaming/internal/auth"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestStreamKeyRoutes_Unauthorized(t *testing.T) {
	t.Parallel()

	router, _ := newTestRouter(t)
	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/stream-keys"},
		{method: http.MethodPost, path: "/api/v1/stream-keys/refresh"},
		{method: http.MethodPost, path: "/api/v1/stream-keys/revoke"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(""))
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestStreamKeyRoutes_CreateRefreshRevoke(t *testing.T) {
	t.Parallel()

	router, redisClient := newTestRouter(t)
	accessToken := mustAccessToken(t)

	createResp := performAuthorizedPost(t, router, accessToken, "/api/v1/stream-keys")
	if createResp.Code != http.StatusOK {
		t.Fatalf("unexpected create status: got %d want %d", createResp.Code, http.StatusOK)
	}
	createBody := decodeBody(t, createResp)
	firstKey := getString(t, createBody, "stream_key")
	if firstKey == "" {
		t.Fatalf("expected non-empty stream_key")
	}

	redisStored, err := redisClient.Get(t.Context(), streamKeyRedisKeyForTest(123)).Result()
	if err != nil {
		t.Fatalf("failed to get stream key from redis: %v", err)
	}
	if redisStored != firstKey {
		t.Fatalf("unexpected redis stream key: got %s want %s", redisStored, firstKey)
	}

	refreshResp := performAuthorizedPost(t, router, accessToken, "/api/v1/stream-keys/refresh")
	if refreshResp.Code != http.StatusOK {
		t.Fatalf("unexpected refresh status: got %d want %d", refreshResp.Code, http.StatusOK)
	}
	refreshBody := decodeBody(t, refreshResp)
	secondKey := getString(t, refreshBody, "stream_key")
	if secondKey == "" || secondKey == firstKey {
		t.Fatalf("expected refreshed stream key to be new")
	}

	revokeResp := performAuthorizedPost(t, router, accessToken, "/api/v1/stream-keys/revoke")
	if revokeResp.Code != http.StatusOK {
		t.Fatalf("unexpected revoke status: got %d want %d", revokeResp.Code, http.StatusOK)
	}
	if _, err := redisClient.Get(t.Context(), streamKeyRedisKeyForTest(123)).Result(); err == nil {
		t.Fatalf("expected stream key to be deleted after revoke")
	}
}

func TestStreamKeyRoutes_CreateIsIdempotentPerUser(t *testing.T) {
	t.Parallel()

	router, redisClient := newTestRouter(t)
	accessToken := mustAccessToken(t)

	firstCreate := performAuthorizedPost(t, router, accessToken, "/api/v1/stream-keys")
	if firstCreate.Code != http.StatusOK {
		t.Fatalf("unexpected first create status: got %d want %d", firstCreate.Code, http.StatusOK)
	}
	firstBody := decodeBody(t, firstCreate)
	firstKey := getString(t, firstBody, "stream_key")

	secondCreate := performAuthorizedPost(t, router, accessToken, "/api/v1/stream-keys")
	if secondCreate.Code != http.StatusOK {
		t.Fatalf("unexpected second create status: got %d want %d", secondCreate.Code, http.StatusOK)
	}
	secondBody := decodeBody(t, secondCreate)
	secondKey := getString(t, secondBody, "stream_key")

	if secondKey != firstKey {
		t.Fatalf("expected same stream key for repeated create: first=%s second=%s", firstKey, secondKey)
	}

	redisStored, err := redisClient.Get(t.Context(), streamKeyRedisKeyForTest(123)).Result()
	if err != nil {
		t.Fatalf("failed to get stream key from redis: %v", err)
	}
	if redisStored != firstKey {
		t.Fatalf("unexpected redis stream key: got %s want %s", redisStored, firstKey)
	}
}

func TestStreamKeyRoutes_RefreshWithoutExistingKey(t *testing.T) {
	t.Parallel()

	router, _ := newTestRouter(t)
	accessToken := mustAccessToken(t)
	rec := performAuthorizedPost(t, router, accessToken, "/api/v1/stream-keys/refresh")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusNotFound)
	}
}

func newTestRouter(t *testing.T) (*gin.Engine, *redis.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	jwtManager, err := authpkg.NewJWTManager("stream-key-secret", "stream-key-issuer")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	r := gin.New()
	v1 := r.Group("/api/v1")
	RegisterRoutes(v1, Deps{
		JWTManager: jwtManager,
		Redis:      redisClient,
	})
	return r, redisClient
}

func mustAccessToken(t *testing.T) string {
	t.Helper()

	jwtManager, err := authpkg.NewJWTManager("stream-key-secret", "stream-key-issuer")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}
	token, err := jwtManager.GenerateToken(123, "streamer@example.com", "streamer", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}
	return token
}

func performAuthorizedPost(t *testing.T, router *gin.Engine, accessToken, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &m); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return m
}

func getString(t *testing.T, body map[string]any, key string) string {
	t.Helper()
	v, ok := body[key].(string)
	if !ok {
		t.Fatalf("expected %s to be string, got %T", key, body[key])
	}
	return v
}

func streamKeyRedisKeyForTest(userID int64) string {
	return fmt.Sprintf("stream:key:%d", userID)
}
