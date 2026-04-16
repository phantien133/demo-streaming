package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Deps struct {
	DB *gorm.DB
}

type HealthResponse struct {
	Status string `json:"status"`
}

func RegisterRoutes(v1 *gin.RouterGroup, deps Deps) {
	// @Summary Health check (DB ping)
	// @Tags system
	// @Produce json
	// @Success 200 {object} HealthResponse
	// @Router /api/v1/health [get]
	v1.GET("/health", func(c *gin.Context) {
		status := "ok"
		if sqlDB, err := deps.DB.DB(); err != nil {
			status = "degraded"
		} else {
			ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
			defer cancel()
			if err := sqlDB.PingContext(ctx); err != nil {
				status = "degraded"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status": status,
		})
	})
}
