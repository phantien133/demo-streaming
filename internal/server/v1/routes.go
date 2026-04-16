package v1

import (
	"demo-streaming/internal/app"
	featureAuth "demo-streaming/internal/server/v1/auth"
	featureHealth "demo-streaming/internal/server/v1/health"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup, container *app.Container) {
	featureHealth.RegisterRoutes(router, featureHealth.Deps{
		DB: container.DB,
	})
	featureAuth.RegisterRoutes(router, featureAuth.Deps{
		JWTManager: container.JWTManager,
		AppConfig:  container.AppConfig,
		Redis:      container.Redis,
	})
}
