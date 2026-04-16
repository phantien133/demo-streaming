package streamkeys

import (
	"context"
	"errors"
	"fmt"

	redisutil "demo-streaming/internal/utils/redis"
	"github.com/redis/go-redis/v9"
)

type RefreshStreamKeyService struct {
	redisUtils *redisutil.RedisUtils
}

func NewRefreshStreamKeyService(redisClient *redis.Client) *RefreshStreamKeyService {
	return &RefreshStreamKeyService{
		redisUtils: redisutil.NewRedisUtils(redisClient),
	}
}

func (s *RefreshStreamKeyService) Execute(ctx context.Context, userID int64) (string, int64, error) {
	key := redisKey(userID)
	if _, err := s.redisUtils.GetString(ctx, key); err != nil {
		if errors.Is(err, redisutil.ErrKeyNotFound) {
			return "", 0, ErrStreamKeyNotFound
		}
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
