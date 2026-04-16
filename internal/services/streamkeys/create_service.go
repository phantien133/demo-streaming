package streamkeys

import (
	"context"
	"errors"
	"fmt"

	redisutil "demo-streaming/internal/utils/redis"
	"github.com/redis/go-redis/v9"
)

type CreateStreamKeyService struct {
	redisUtils *redisutil.RedisUtils
}

func NewCreateStreamKeyService(redisClient *redis.Client) *CreateStreamKeyService {
	return &CreateStreamKeyService{
		redisUtils: redisutil.NewRedisUtils(redisClient),
	}
}

func (s *CreateStreamKeyService) Execute(ctx context.Context, userID int64) (string, int64, error) {
	key := redisKey(userID)
	existingKey, err := s.redisUtils.GetString(ctx, key)
	if err == nil {
		return existingKey, int64(defaultStreamKeyTTL.Seconds()), nil
	}
	if !errors.Is(err, redisutil.ErrKeyNotFound) {
		return "", 0, fmt.Errorf("%w: failed to check stream key", ErrStreamKeyStore)
	}

	streamKey, err := generateStreamKey()
	if err != nil {
		return "", 0, err
	}
	if err := s.redisUtils.SetString(ctx, key, streamKey, defaultStreamKeyTTL); err != nil {
		return "", 0, fmt.Errorf("%w: failed to store stream key", ErrStreamKeyStore)
	}
	return streamKey, int64(defaultStreamKeyTTL.Seconds()), nil
}
