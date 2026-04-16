package auth

import (
	"demo-streaming/internal/auth"
	"demo-streaming/internal/config"
	"demo-streaming/internal/middleware"
	authservice "demo-streaming/internal/services/auth"
	redisutil "demo-streaming/internal/utils/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Deps struct {
	JWTManager *auth.JWTManager
	AppConfig  config.AppConfig
	Redis      *redis.Client
	DB         *gorm.DB
	LoginService authservice.LoginService
}

func RegisterRoutes(v1 *gin.RouterGroup, deps Deps) {
	loginService := deps.LoginService
	if loginService == nil {
		loginService = authservice.NewGormLoginService(deps.DB)
	}

	handler := &Handler{
		JWTManager: deps.JWTManager,
		AppConfig:  deps.AppConfig,
		Redis:      deps.Redis,
		RedisUtils: redisutil.NewRedisUtils(deps.Redis),
		LoginService: loginService,
	}

	v1.POST("/auth/token", handler.CreateToken)
	v1.POST("/auth/refresh", handler.Refresh)
	v1.POST("/auth/revoke", handler.Revoke)

	authGroup := v1.Group("/auth")
	authGroup.Use(middleware.JWTAuth(deps.JWTManager))
	authGroup.GET("/me", handler.Me)
}
