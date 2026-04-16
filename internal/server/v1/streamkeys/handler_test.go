package streamkeys

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authpkg "demo-streaming/internal/auth"
	"demo-streaming/internal/middleware"
	streamkeysservice "demo-streaming/internal/services/streamkeys"
	"github.com/gin-gonic/gin"
)

func TestHandlerCreate_ErrorMapping(t *testing.T) {
	t.Parallel()

	h := &Handler{
		CreateService: createServiceFunc(func(context.Context, int64, *int64) (string, int64, error) {
			return "", 0, errors.New("boom")
		}),
	}

	rec := performWithClaims(t, h.Create)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestHandlerRefresh_ErrorMapping(t *testing.T) {
	t.Parallel()

	t.Run("returns 404 for stream key not found", func(t *testing.T) {
		t.Parallel()
		h := &Handler{
			RefreshService: refreshServiceFunc(func(context.Context, int64) (string, int64, error) {
				return "", 0, streamkeysservice.ErrStreamKeyNotFound
			}),
		}

		rec := performWithClaims(t, h.Refresh)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusNotFound)
		}
	})

	t.Run("returns 500 for generic refresh error", func(t *testing.T) {
		t.Parallel()
		h := &Handler{
			RefreshService: refreshServiceFunc(func(context.Context, int64) (string, int64, error) {
				return "", 0, errors.New("boom")
			}),
		}

		rec := performWithClaims(t, h.Refresh)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusInternalServerError)
		}
	})
}

func TestHandlerRevoke_ErrorMapping(t *testing.T) {
	t.Parallel()

	h := &Handler{
		RevokeService: revokeServiceFunc(func(context.Context, int64) error {
			return errors.New("boom")
		}),
	}

	rec := performWithClaims(t, h.Revoke)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusInternalServerError)
	}
}

func performWithClaims(t *testing.T, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/test", func(c *gin.Context) {
		c.Set(middleware.AuthClaimsContextKey, &authpkg.Claims{UserID: 123})
		handler(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(""))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

type createServiceFunc func(ctx context.Context, userID int64, expiresInSeconds *int64) (string, int64, error)

func (f createServiceFunc) Execute(ctx context.Context, userID int64, expiresInSeconds *int64) (string, int64, error) {
	return f(ctx, userID, expiresInSeconds)
}

type refreshServiceFunc func(ctx context.Context, userID int64) (string, int64, error)

func (f refreshServiceFunc) Execute(ctx context.Context, userID int64) (string, int64, error) {
	return f(ctx, userID)
}

type revokeServiceFunc func(ctx context.Context, userID int64) error

func (f revokeServiceFunc) Execute(ctx context.Context, userID int64) error {
	return f(ctx, userID)
}
