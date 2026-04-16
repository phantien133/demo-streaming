package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	internalAuth "demo-streaming/internal/auth"
	"demo-streaming/internal/config"
	"github.com/gin-gonic/gin"
)

func TestRegisterRoutes_EndpointsAreReachable(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)
	router, jwtManager := newTestRouter(t)

	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "token route registered and validates payload",
			method:     http.MethodPost,
			target:     "/api/v1/auth/token",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "refresh route registered and validates payload",
			method:     http.MethodPost,
			target:     "/api/v1/auth/refresh",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "revoke route registered and validates payload",
			method:     http.MethodPost,
			target:     "/api/v1/auth/revoke",
			body:       `{}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "me route is protected without auth",
			method:     http.MethodGet,
			target:     "/api/v1/auth/me",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "me route allows valid bearer token",
			method:     http.MethodGet,
			target:     "/api/v1/auth/me",
			authHeader: mustAccessToken(t, jwtManager),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var bodyReader *strings.Reader
			if tt.body == "" {
				bodyReader = strings.NewReader("")
			} else {
				bodyReader = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.target, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", "Bearer "+tt.authHeader)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var got map[string]any
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if got["email"] != "user@example.com" {
					t.Fatalf("unexpected email in response: %v", got["email"])
				}
			}
		})
	}
}

func newTestRouter(t *testing.T) (*gin.Engine, *internalAuth.JWTManager) {
	t.Helper()

	jwtManager, err := internalAuth.NewJWTManager("test-secret", "test-issuer")
	if err != nil {
		t.Fatalf("failed to create jwt manager: %v", err)
	}

	r := gin.New()
	v1 := r.Group("/api/v1")
	RegisterRoutes(v1, Deps{
		JWTManager: jwtManager,
		AppConfig: config.AppConfig{
			JWTAccessTokenTTLSeconds:  3600,
			JWTRefreshTokenTTLSeconds: 7200,
		},
	})

	return r, jwtManager
}

func mustAccessToken(t *testing.T, jwtManager *internalAuth.JWTManager) string {
	t.Helper()

	token, err := jwtManager.GenerateToken(123, "user@example.com", "end_user", time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	return token
}
