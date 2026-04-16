package main

import (
	"context"
	"errors"
	"log"
	"os/signal"
	"syscall"

	"demo-streaming/internal/app"
	"demo-streaming/internal/config"
	transcodeservice "demo-streaming/internal/services/transcode"
)

func main() {
	config.LoadDotEnv()
	systemCfg := config.LoadSystemConfig()
	appCfg := config.LoadAppConfig()

	container, err := app.NewContainer(systemCfg, appCfg)
	if err != nil {
		log.Fatalf("failed to initialize transcode container: %v", err)
	}
	defer func() {
		if err := container.Close(); err != nil {
			log.Printf("failed to close transcode container: %v", err)
		}
	}()

	queue := transcodeservice.NewRedisQueue(container.Redis, appCfg.TranscodeQueueKey)
	runner := transcodeservice.NewRunner(appCfg)
	recorder := transcodeservice.NewGormMetadataStore(container.DB)
	worker := transcodeservice.NewWorker(queue, runner, recorder)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("transcode worker started queue=%s enabled=%t", appCfg.TranscodeQueueKey, appCfg.TranscodeEnabled)
	if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("transcode worker stopped with error: %v", err)
	}
	log.Println("transcode worker stopped")
}
