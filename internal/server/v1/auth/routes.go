package auth

import (
	"demo-streaming/internal/auth"
	"demo-streaming/internal/config"
	"demo-streaming/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	JWTManager *auth.JWTManager
	AppConfig  config.AppConfig
	Redis      *redis.Client
}

func RegisterRoutes(v1 *gin.RouterGroup, deps Deps) {
	handler := &Handler{
		JWTManager: deps.JWTManager,
		AppConfig:  deps.AppConfig,
		Redis:      deps.Redis,
	}

	v1.POST("/auth/token", handler.CreateToken)
	v1.POST("/auth/refresh", handler.Refresh)
	v1.POST("/auth/revoke", handler.Revoke)

	authGroup := v1.Group("/auth")
	authGroup.Use(middleware.JWTAuth(deps.JWTManager))
	authGroup.GET("/me", handler.Me)
}
