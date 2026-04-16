package server

import (
	"net/http"

	"demo-streaming/docs"
	"demo-streaming/internal/app"
	apiv1 "demo-streaming/internal/server/v1"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

func NewRouter(container *app.Container) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	// @Summary Health check (gateway)
	// @Tags system
	// @Produce json
	// @Success 200 {object} map[string]string
	// @Router /healthz [get]
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	v1 := router.Group("/api/v1")
	apiv1.Register(v1, container)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "api-gateway",
		})
	})

	return router
}
