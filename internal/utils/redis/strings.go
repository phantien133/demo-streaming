package redis

import (
	"context"
	"errors"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var ErrKeyNotFound = errors.New("redis key not found")

type RedisUtils struct {
	client *goredis.Client
}

func NewRedisUtils(client *goredis.Client) *RedisUtils {
	return &RedisUtils{client: client}
}

func (r *RedisUtils) GetString(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return "", ErrKeyNotFound
		}
		return "", err
	}
	return value, nil
}

func (r *RedisUtils) SetString(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisUtils) DeleteKey(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
