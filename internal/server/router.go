package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "api-gateway",
		})
	})

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			status := "ok"
			if sqlDB, err := db.DB(); err != nil {
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

	return router
}
