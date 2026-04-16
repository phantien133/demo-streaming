package streamkeys

import (
	"errors"
	"net/http"

	"demo-streaming/internal/auth"
	"demo-streaming/internal/middleware"
	streamkeysservice "demo-streaming/internal/services/streamkeys"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	CreateService  streamkeysservice.CreateService
	RefreshService streamkeysservice.RefreshService
	RevokeService  streamkeysservice.RevokeService
}

func (h *Handler) Create(c *gin.Context) {
	claims, ok := authClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	streamKey, expiresIn, err := h.CreateService.Execute(c.Request.Context(), claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate stream key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stream_key": streamKey,
		"expires_in": expiresIn,
	})
}

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
