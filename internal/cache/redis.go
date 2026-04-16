package cache

import (
	"context"
	"fmt"
	"time"

	"demo-streaming/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg config.SystemConfig) (*redis.Client, func() error, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, nil, err
	}

	return client, client.Close, nil
}
