package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authpkg "demo-streaming/internal/auth"
	"demo-streaming/internal/config"
	"demo-streaming/internal/middleware"
	redisutil "demo-streaming/internal/utils/redis"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestHandlerCreateToken(t *testing.T) {
	t.Parallel()

	t.Run("returns bad request for invalid payload", func(t *testing.T) {
		t.Parallel()

		h, _, _ := newTestHandler(t)
		rec := performJSONRequest(t, h.CreateToken, http.MethodPost, "/auth/token", `{}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("returns token pair and stores refresh state", func(t *testing.T) {
		t.Parallel()

		h, _, _ := newTestHandler(t)
		rec := performJSONRequest(t, h.CreateToken, http.MethodPost, "/auth/token", `{"user_id":123,"email":"user@example.com"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
		}

		body := decodeJSON(t, rec)
		refreshToken := asString(t, body["refresh_token"])
		claims := parseClaims(t, h.JWTManager, refreshToken)

		userTokenID, err := h.Redis.Get(t.Context(), userRefreshKey(123)).Result()
		if err != nil {
			t.Fatalf("failed to get user refresh key: %v", err)
		}
		if userTokenID != claims.ID {
			t.Fatalf("unexpected token id mapping: got %s want %s", userTokenID, claims.ID)
		}

		if _, err := h.Redis.Get(t.Context(), refreshKey(claims.ID)).Result(); err != nil {
			t.Fatalf("expected refresh key to be stored: %v", err)
		}
	})
}

func TestHandlerRefresh(t *testing.T) {
	t.Parallel()

	t.Run("returns unauthorized when refresh token is revoked", func(t *testing.T) {
		t.Parallel()

		h, _, _ := newTestHandler(t)
		refreshToken, _, err := h.JWTManager.GenerateRefreshToken(123, "user@example.com", "end_user", time.Hour)
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		rec := performJSONRequest(t, h.Refresh, http.MethodPost, "/auth/refresh", `{"refresh_token":"`+refreshToken+`"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("rotates refresh token successfully", func(t *testing.T) {
		t.Parallel()

		h, _, _ := newTestHandler(t)
		refreshToken, oldTokenID, err := h.JWTManager.GenerateRefreshToken(123, "user@example.com", "end_user", time.Hour)
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		if err := h.Redis.Set(t.Context(), refreshKey(oldTokenID), time.Now().Unix(), time.Hour).Err(); err != nil {
			t.Fatalf("failed to seed refresh key: %v", err)
		}
		if err := h.Redis.Set(t.Context(), userRefreshKey(123), oldTokenID, time.Hour).Err(); err != nil {
			t.Fatalf("failed to seed user refresh key: %v", err)
		}

		rec := performJSONRequest(t, h.Refresh, http.MethodPost, "/auth/refresh", `{"refresh_token":"`+refreshToken+`"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
		}

		body := decodeJSON(t, rec)
		newRefreshToken := asString(t, body["refresh_token"])
		newClaims := parseClaims(t, h.JWTManager, newRefreshToken)

		if _, err := h.Redis.Get(t.Context(), refreshKey(oldTokenID)).Result(); err == nil {
			t.Fatalf("expected old refresh key to be removed")
		}
		if _, err := h.Redis.Get(t.Context(), refreshKey(newClaims.ID)).Result(); err != nil {
			t.Fatalf("expected new refresh key to exist: %v", err)
		}
	})
}

func TestHandlerRevoke(t *testing.T) {
	t.Parallel()

	h, _, _ := newTestHandler(t)
	refreshToken, tokenID, err := h.JWTManager.GenerateRefreshToken(123, "user@example.com", "end_user", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	if err := h.Redis.Set(t.Context(), refreshKey(tokenID), time.Now().Unix(), time.Hour).Err(); err != nil {
		t.Fatalf("failed to seed refresh key: %v", err)
	}
	if err := h.Redis.Set(t.Context(), userRefreshKey(123), tokenID, time.Hour).Err(); err != nil {
		t.Fatalf("failed to seed user refresh key: %v", err)
	}

	rec := performJSONRequest(t, h.Revoke, http.MethodPost, "/auth/revoke", `{"refresh_token":"`+refreshToken+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}

	if _, err := h.Redis.Get(t.Context(), refreshKey(tokenID)).Result(); err == nil {
		t.Fatalf("expected refresh key to be deleted")
	}
	if _, err := h.Redis.Get(t.Context(), userRefreshKey(123)).Result(); err == nil {
		t.Fatalf("expected user refresh key to be deleted")
	}
}

func TestHandlerMe(t *testing.T) {
	t.Parallel()

	t.Run("returns unauthorized when claims are missing", func(t *testing.T) {
		t.Parallel()
		h, _, _ := newTestHandler(t)
		rec := performRequest(t, func(r *gin.Engine) {
			r.GET("/auth/me", h.Me)
		}, http.MethodGet, "/auth/me")
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("returns current claims payload", func(t *testing.T) {
		t.Parallel()
		h, _, _ := newTestHandler(t)
		rec := performRequest(t, func(r *gin.Engine) {
			r.GET("/auth/me", func(c *gin.Context) {
				c.Set(middleware.AuthClaimsContextKey, &authpkg.Claims{
					UserID: 321,
					Email:  "me@example.com",
					Role:   "end_user",
				})
				h.Me(c)
			})
		}, http.MethodGet, "/auth/me")
		if rec.Code != http.StatusOK {
			t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
		}

		body := decodeJSON(t, rec)
		if asString(t, body["email"]) != "me@example.com" {
			t.Fatalf("unexpected email: %v", body["email"])
		}
	})
}

func newTestHandler(t *testing.T) (*Handler, *miniredis.Miniredis, *redis.Client) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	jwtManager, err := authpkg.NewJWTManager("test-secret", "test-issuer")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	return &Handler{
		JWTManager: jwtManager,
		AppConfig: config.AppConfig{
			JWTAccessTokenTTLSeconds:  3600,
			JWTRefreshTokenTTLSeconds: 7200,
		},
		Redis:      redisClient,
		RedisUtils: redisutil.NewRedisUtils(redisClient),
	}, mr, redisClient
}

func performJSONRequest(
	t *testing.T,
	handler gin.HandlerFunc,
	method, target, body string,
) *httptest.ResponseRecorder {
	t.Helper()
	return performRequest(t, func(r *gin.Engine) {
		r.Handle(method, target, handler)
	}, method, target, body)
}

func performRequest(
	t *testing.T,
	setup func(*gin.Engine),
	method, target string,
	body ...string,
) *httptest.ResponseRecorder {
	t.Helper()

	r := gin.New()
	setup(r)

	payload := ""
	if len(body) > 0 {
		payload = body[0]
	}

	req := httptest.NewRequest(method, target, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return body
}

func parseClaims(t *testing.T, jwtManager *authpkg.JWTManager, token string) *authpkg.Claims {
	t.Helper()
	claims, err := jwtManager.ParseToken(token)
	if err != nil {
		t.Fatalf("failed to parse token claims: %v", err)
	}
	return claims
}

func asString(t *testing.T, v any) string {
	t.Helper()
	s, ok := v.(string)
	if !ok {
		t.Fatalf("expected string, got %T", v)
	}
	return s
}
