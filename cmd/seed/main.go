package main

import (
	"context"
	"flag"
	"log"
	"time"

	"demo-streaming/internal/app"
	"demo-streaming/internal/config"
	"demo-streaming/internal/services/seed"
)

func main() {
	config.LoadDotEnv()
	systemCfg := config.LoadSystemConfig()
	appCfg := config.LoadAppConfig()

	var (
		reset           = flag.Bool("reset", false, "truncate tables before seeding (destructive)")
		users           = flag.Int("users", 3, "number of streamer users to seed")
		usersFile       = flag.String("users-file", "internal/services/seed/fixtures/users.json", "path to users JSON fixture file")
		streamKeysPerUser = flag.Int("stream-keys-per-user", 1, "stream keys per user")
		sessionsPerUser = flag.Int("sessions-per-user", 1, "publish sessions per user")
		timeout         = flag.Duration("timeout", 10*time.Second, "db operation timeout")
	)
	flag.Parse()

	container, err := app.NewContainer(systemCfg, appCfg)
	if err != nil {
		log.Fatalf("failed to initialize app container: %v", err)
	}
	defer func() {
		if err := container.Close(); err != nil {
			log.Printf("failed to close app container: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	svc := seed.NewSeedLocalGormService(container.DB, container.AppConfig)
	out, err := svc.Execute(ctx, seed.SeedLocalInput{
		Reset:            *reset,
		UserCount:        *users,
		StreamKeysPerUser: *streamKeysPerUser,
		SessionsPerUser:  *sessionsPerUser,
		UsersFile:        *usersFile,
	})
	if err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Printf("seed completed: users=%d providers=%d stream_keys=%d publish_sessions=%d",
		out.UsersCreated, out.MediaProvidersEnsured, out.StreamKeysCreated, out.PublishSessionsCreated,
	)
}

