package streamkeys

import (
	"context"
	"fmt"

	redisutil "demo-streaming/internal/utils/redis"
	"github.com/redis/go-redis/v9"
)

type RevokeStreamKeyService struct {
	redisUtils *redisutil.RedisUtils
}

func NewRevokeStreamKeyService(redisClient *redis.Client) *RevokeStreamKeyService {
	return &RevokeStreamKeyService{
		redisUtils: redisutil.NewRedisUtils(redisClient),
	}
}

func (s *RevokeStreamKeyService) Execute(ctx context.Context, userID int64) error {
	if err := s.redisUtils.DeleteKey(ctx, redisKey(userID)); err != nil {
		return fmt.Errorf("%w: failed to revoke stream key", ErrStreamKeyStore)
	}
	return nil
}
