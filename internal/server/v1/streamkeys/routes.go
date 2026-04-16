package streamkeys

import (
	"demo-streaming/internal/auth"
	"demo-streaming/internal/config"
	"demo-streaming/internal/middleware"
	streamkeysservice "demo-streaming/internal/services/streamkeys"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Deps struct {
	JWTManager *auth.JWTManager
	Redis      *redis.Client
	DB         *gorm.DB
	AppConfig  config.AppConfig
}

func RegisterRoutes(v1 *gin.RouterGroup, deps Deps) {
	handler := &Handler{
		DB:            deps.DB,
		AppConfig:     deps.AppConfig,
		CreateService:  streamkeysservice.NewCreateStreamKeyService(deps.DB, deps.Redis),
		RefreshService: streamkeysservice.NewRefreshStreamKeyService(deps.DB, deps.Redis),
		RevokeService:  streamkeysservice.NewRevokeStreamKeyService(deps.DB, deps.Redis),
	}

	group := v1.Group("/stream-keys")
	group.Use(middleware.JWTAuth(deps.JWTManager))
	group.POST("", handler.Create)
	group.POST("/refresh", handler.Refresh)
	group.POST("/revoke", handler.Revoke)
}
