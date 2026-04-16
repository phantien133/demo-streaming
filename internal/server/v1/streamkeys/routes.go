package streamkeys

import (
	"demo-streaming/internal/auth"
	"demo-streaming/internal/middleware"
	streamkeysservice "demo-streaming/internal/services/streamkeys"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Deps struct {
	JWTManager *auth.JWTManager
	Redis      *redis.Client
}

func RegisterRoutes(v1 *gin.RouterGroup, deps Deps) {
	handler := &Handler{
		CreateService:  streamkeysservice.NewCreateStreamKeyService(deps.Redis),
		RefreshService: streamkeysservice.NewRefreshStreamKeyService(deps.Redis),
		RevokeService:  streamkeysservice.NewRevokeStreamKeyService(deps.Redis),
	}

	group := v1.Group("/stream-keys")
	group.Use(middleware.JWTAuth(deps.JWTManager))
	group.POST("", handler.Create)
	group.POST("/refresh", handler.Refresh)
	group.POST("/revoke", handler.Revoke)
}
