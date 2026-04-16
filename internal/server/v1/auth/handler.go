package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"demo-streaming/internal/auth"
	"demo-streaming/internal/config"
	"demo-streaming/internal/middleware"
	authservice "demo-streaming/internal/services/auth"
	redisutil "demo-streaming/internal/utils/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	JWTManager   *auth.JWTManager
	AppConfig    config.AppConfig
	Redis        *redis.Client
	RedisUtils   *redisutil.RedisUtils
	LoginService authservice.LoginService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateTokenRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RevokeTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type TokenPairResponse struct {
	AccessToken           string `json:"access_token"`
	RefreshToken          string `json:"refresh_token"`
	TokenType             string `json:"token_type"`
	AccessTokenExpiresIn  int64  `json:"access_token_expires_in"`
	RefreshTokenExpiresIn int64  `json:"refresh_token_expires_in"`
}

type RevokeResponse struct {
	Status string `json:"status"`
}

type MeResponse struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Issuer string `json:"issuer"`
}

const (
	defaultEndUserRole  = "end_user"
	refreshRedisKeyPref = "auth:refresh:"
	userRefreshKeyPref  = "auth:user_refresh:"
)

// CreateToken issues access + refresh tokens.
//
// @Summary Login and create token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body CreateTokenRequest true "login credentials"
// @Success 200 {object} TokenPairResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/token [post]
func (h *Handler) CreateToken(c *gin.Context) {
	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginOut, err := h.LoginService.Execute(c.Request.Context(), authservice.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, authservice.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	ttl := h.accessTokenTTL()
	refreshTTL := h.refreshTokenTTL()

	accessToken, err := h.JWTManager.GenerateToken(loginOut.UserID, loginOut.Email, defaultEndUserRole, ttl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	refreshToken, refreshTokenID, err := h.JWTManager.GenerateRefreshToken(loginOut.UserID, loginOut.Email, defaultEndUserRole, refreshTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	ctx := c.Request.Context()
	providedAt := time.Now().Unix()
	if err := h.rotateUserRefreshToken(ctx, loginOut.UserID, "", refreshTokenID, providedAt, refreshTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokenPairResponse(accessToken, refreshToken, ttl, refreshTTL))
}

// Refresh rotates refresh token and returns a new token pair.
//
// @Summary Refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "refresh token"
// @Success 200 {object} TokenPairResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.parseAndValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if _, err := h.RedisUtils.GetString(ctx, refreshKey(claims.ID)); err != nil {
		if errors.Is(err, redisutil.ErrKeyNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token revoked or expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check refresh token state"})
		return
	}

	accessTTL := h.accessTokenTTL()
	refreshTTL := h.refreshTokenTTL()
	newAccessToken, err := h.JWTManager.GenerateToken(claims.UserID, claims.Email, claims.Role, accessTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate access token"})
		return
	}

	newRefreshToken, newRefreshTokenID, err := h.JWTManager.GenerateRefreshToken(claims.UserID, claims.Email, claims.Role, refreshTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate refresh token"})
		return
	}

	if err := h.rotateUserRefreshToken(ctx, claims.UserID, claims.ID, newRefreshTokenID, time.Now().Unix(), refreshTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store rotated refresh token"})
		return
	}

	c.JSON(http.StatusOK, tokenPairResponse(newAccessToken, newRefreshToken, accessTTL, refreshTTL))
}

// Revoke invalidates a refresh token immediately.
//
// @Summary Revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RevokeTokenRequest true "refresh token"
// @Success 200 {object} RevokeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/revoke [post]
func (h *Handler) Revoke(c *gin.Context) {
	var req RevokeTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.parseAndValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	if err := h.RedisUtils.DeleteKey(ctx, refreshKey(claims.ID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke refresh token"})
		return
	}
	userKey := userRefreshKey(claims.UserID)
	currentTokenID, err := h.RedisUtils.GetString(ctx, userKey)
	if err == nil && currentTokenID == claims.ID {
		if err := h.RedisUtils.DeleteKey(ctx, userKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke refresh token"})
			return
		}
	} else if err != nil && !errors.Is(err, redisutil.ErrKeyNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "revoked"})
}

// Me returns JWT claims for the current user.
//
// @Summary Get current user claims
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} MeResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	rawClaims, ok := c.Get(middleware.AuthClaimsContextKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth claims"})
		return
	}

	claims, ok := rawClaims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth claims"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
		"issuer":  claims.Issuer,
	})
}

func refreshKey(tokenID string) string {
	return refreshRedisKeyPref + tokenID
}

func userRefreshKey(userID int64) string {
	return fmt.Sprintf("%s%d", userRefreshKeyPref, userID)
}

func (h *Handler) rotateUserRefreshToken(
	ctx context.Context,
	userID int64,
	oldTokenID string,
	newTokenID string,
	providedAt int64,
	refreshTTL time.Duration,
) error {
	userKey := userRefreshKey(userID)
	if oldTokenID == "" {
		currentTokenID, err := h.RedisUtils.GetString(ctx, userKey)
		if err == nil && currentTokenID != "" {
			oldTokenID = currentTokenID
		} else if err != nil && !errors.Is(err, redisutil.ErrKeyNotFound) {
			return err
		}
	}

	pipe := h.Redis.TxPipeline()
	if oldTokenID != "" {
		pipe.Del(ctx, refreshKey(oldTokenID))
	}
	pipe.Set(ctx, refreshKey(newTokenID), providedAt, refreshTTL)
	pipe.Set(ctx, userKey, newTokenID, refreshTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (h *Handler) accessTokenTTL() time.Duration {
	ttl := time.Duration(h.AppConfig.JWTAccessTokenTTLSeconds) * time.Second
	if ttl <= 0 {
		return time.Hour
	}
	return ttl
}

func (h *Handler) refreshTokenTTL() time.Duration {
	ttl := time.Duration(h.AppConfig.JWTRefreshTokenTTLSeconds) * time.Second
	if ttl <= 0 {
		return 7 * 24 * time.Hour
	}
	return ttl
}

func (h *Handler) parseAndValidateRefreshToken(token string) (*auth.Claims, error) {
	claims, err := h.JWTManager.ParseToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if !strings.EqualFold(claims.TokenType, "refresh") || claims.ID == "" {
		return nil, fmt.Errorf("invalid refresh token type")
	}
	return claims, nil
}

func tokenPairResponse(accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) gin.H {
	return gin.H{
		"access_token":             accessToken,
		"refresh_token":            refreshToken,
		"token_type":               "Bearer",
		"access_token_expires_in":  int64(accessTTL.Seconds()),
		"refresh_token_expires_in": int64(refreshTTL.Seconds()),
	}
}
