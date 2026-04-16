package playback

import (
	"demo-streaming/internal/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Deps struct {
	DB        *gorm.DB
	AppConfig config.AppConfig
}

func RegisterRoutes(v1 *gin.RouterGroup, deps Deps) {
	h := &Handler{
		DB:        deps.DB,
		AppConfig: deps.AppConfig,
	}

	group := v1.Group("/playback")
	group.GET("/:playback_id/index.m3u8", h.Playlist)
	group.GET("/:playback_id/segments/:name", h.Segment)
}

