package streamkeys

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"demo-streaming/internal/auth"
	"demo-streaming/internal/config"
	"demo-streaming/internal/database"
	"demo-streaming/internal/middleware"
	streamkeysservice "demo-streaming/internal/services/streamkeys"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	DB        *gorm.DB
	AppConfig config.AppConfig

	CreateService  streamkeysservice.CreateService
	RefreshService streamkeysservice.RefreshService
	RevokeService  streamkeysservice.RevokeService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StreamKeyResponse struct {
	StreamKey string `json:"stream_key"`
	ExpiresIn int64  `json:"expires_in"`
	Ingest    IngestInfo `json:"ingest"`
}

type IngestInfo struct {
	Provider string `json:"provider"`
	RTMPURL  string `json:"rtmp_url"`
}

type RevokeResponse struct {
	Status string `json:"status"`
}

type CreateRequest struct {
	// Optional: if provided (>0), key will expire after this many seconds.
	ExpiresIn *int64 `json:"expires_in"`
}

// Create returns a stream key for current user (idempotent).
//
// @Summary Create stream key
// @Tags stream-keys
// @Security BearerAuth
// @Produce json
// @Accept json
// @Success 200 {object} StreamKeyResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream-keys [post]
func (h *Handler) Create(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	var req CreateRequest
	_ = c.ShouldBindJSON(&req)

	streamKey, expiresIn, err := h.CreateService.Execute(c.Request.Context(), claims.UserID, req.ExpiresIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate stream key"})
		return
	}

	rtmpURL := ""
	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "media provider not configured"})
		return
	}

	var provider database.MediaProvider
	if err := h.DB.WithContext(c.Request.Context()).Where("code = ?", "srs").First(&provider).Error; err == nil {
		var cfg struct {
			RTMPBaseURL string `json:"rtmp_base_url"`
		}
		_ = json.Unmarshal(provider.Config, &cfg)
		rtmpURL = strings.TrimSpace(cfg.RTMPBaseURL)
	}
	if strings.TrimSpace(rtmpURL) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "media provider misconfigured"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stream_key": streamKey,
		"expires_in": expiresIn,
		"ingest": gin.H{
			"provider": "srs",
			"rtmp_url": rtmpURL,
		},
	})
}

// Refresh rotates the current stream key.
//
// @Summary Refresh stream key
// @Tags stream-keys
// @Security BearerAuth
// @Produce json
// @Success 200 {object} StreamKeyResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream-keys/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	streamKey, expiresIn, err := h.RefreshService.Execute(c.Request.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, streamkeysservice.ErrStreamKeyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "stream key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh stream key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stream_key": streamKey,
		"expires_in": expiresIn,
	})
}

// Revoke removes the current stream key.
//
// @Summary Revoke stream key
// @Tags stream-keys
// @Security BearerAuth
// @Produce json
// @Success 200 {object} RevokeResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/stream-keys/revoke [post]
func (h *Handler) Revoke(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	if err := h.RevokeService.Execute(c.Request.Context(), claims.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke stream key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

func authClaimsFromContext(c *gin.Context) (*auth.Claims, bool) {
	rawClaims, ok := c.Get(middleware.AuthClaimsContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := rawClaims.(*auth.Claims)
	return claims, ok
}
