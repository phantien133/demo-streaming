package streampublishsessions

import (
	"demo-streaming/internal/auth"
	"demo-streaming/internal/config"
	"demo-streaming/internal/middleware"
	streampublishsessionsservice "demo-streaming/internal/services/streampublishsessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Deps struct {
	DB           *gorm.DB
	JWTManager   *auth.JWTManager
	AppConfig    config.AppConfig
	ListService  streampublishsessionsservice.ListService
	CreateService streampublishsessionsservice.CreateService
	StartService streampublishsessionsservice.StartService
	StopLiveService streampublishsessionsservice.StopLiveService
}

func RegisterRoutes(v1 *gin.RouterGroup, deps Deps) {
	createService := deps.CreateService
	if createService == nil {
		createService = streampublishsessionsservice.NewGormCreateService(deps.DB, deps.AppConfig)
	}

	startService := deps.StartService
	if startService == nil {
		startService = streampublishsessionsservice.NewGormStartService(deps.DB)
	}

	stopLiveService := deps.StopLiveService
	if stopLiveService == nil {
		stopLiveService = streampublishsessionsservice.NewGormStopLiveService(deps.DB)
	}

	listService := deps.ListService
	if listService == nil {
		listService = streampublishsessionsservice.NewGormListService(deps.DB)
	}

	handler := &Handler{
		CreateService: createService,
		ListService: listService,
		StartService: startService,
		StopLiveService: stopLiveService,
	}

	group := v1.Group("/stream-publish-sessions")
	group.Use(middleware.JWTAuth(deps.JWTManager))
	group.GET("", handler.List)
	group.POST("", handler.Create)
	group.POST("/:id/start", handler.Start)
	group.POST("/:id/stop", handler.Stop)
	group.POST("/stop-all", handler.StopAll)
}
