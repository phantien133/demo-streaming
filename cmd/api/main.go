package main

import (
	"log"

	"demo-streaming/internal/app"
	"demo-streaming/internal/config"
	"demo-streaming/internal/server"
)

// @title demo-streaming API
// @version 1.0
// @description Go + Gin baseline for a learning-focused streaming system.
//
// @BasePath /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	config.LoadDotEnv()
	systemCfg := config.LoadSystemConfig()
	appCfg := config.LoadAppConfig()

	container, err := app.NewContainer(systemCfg, appCfg)
	if err != nil {
		log.Fatalf("failed to initialize app container: %v", err)
	}
	defer func() {
		if err := container.Close(); err != nil {
			log.Printf("failed to close app container: %v", err)
		}
	}()

	router := server.NewRouter(container)
	addr := ":" + container.SystemConfig.Port

	log.Printf("api server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("failed to run api server: %v", err)
	}
}
