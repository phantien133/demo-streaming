package v1

import (
	"demo-streaming/internal/app"
	"demo-streaming/internal/server/v1/auth"
	"demo-streaming/internal/server/v1/health"
	"demo-streaming/internal/server/v1/streamkeys"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup, container *app.Container) {
	health.RegisterRoutes(router, health.Deps{
		DB: container.DB,
	})
	auth.RegisterRoutes(router, auth.Deps{
		JWTManager: container.JWTManager,
		AppConfig:  container.AppConfig,
		Redis:      container.Redis,
	})
	streamkeys.RegisterRoutes(router, streamkeys.Deps{
		JWTManager: container.JWTManager,
		Redis:      container.Redis,
	})
}
