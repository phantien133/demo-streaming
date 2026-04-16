package server

import (
	"net/http"

	"demo-streaming/internal/app"
	apiV1 "demo-streaming/internal/server/v1"

	"github.com/gin-gonic/gin"
)

func NewRouter(container *app.Container) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	v1 := router.Group("/api/v1")
	apiV1.Register(v1, container)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "api-gateway",
		})
	})

	return router
}
