package mediawebhooks

import (
	"demo-streaming/internal/config"
	mediawebhooksservice "demo-streaming/internal/services/mediawebhooks"
	transcodeservice "demo-streaming/internal/services/transcode"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Deps struct {
	DB        *gorm.DB
	AppConfig config.AppConfig
	Redis     *redis.Client

	SRSOnPublishService mediawebhooksservice.SRSOnPublishService
}

func RegisterRoutes(v1 *gin.RouterGroup, deps Deps) {
	svc := deps.SRSOnPublishService
	if svc == nil {
		queue := transcodeservice.NewRedisQueue(deps.Redis, deps.AppConfig.TranscodeQueueKey)
		svc = mediawebhooksservice.NewGormSRSOnPublishService(deps.DB, queue)
	}

	h := &Handler{
		AppConfig:           deps.AppConfig,
		SRSOnPublishService: svc,
	}

	group := v1.Group("/media/webhooks/srs")
	group.POST("/on-publish", h.OnPublish)
}

