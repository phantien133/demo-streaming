package streamkeys

import (
	"context"
	"fmt"
	"time"

	"demo-streaming/internal/database"
	redisutil "demo-streaming/internal/utils/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RevokeStreamKeyService struct {
	db        *gorm.DB
	redisUtils *redisutil.RedisUtils
}

func NewRevokeStreamKeyService(db *gorm.DB, redisClient *redis.Client) *RevokeStreamKeyService {
	return &RevokeStreamKeyService{
		db:        db,
		redisUtils: redisutil.NewRedisUtils(redisClient),
	}
}

func (s *RevokeStreamKeyService) Execute(ctx context.Context, userID int64) error {
	if s.db == nil {
		return fmt.Errorf("%w: db not configured", ErrStreamKeyStore)
	}

	if err := s.redisUtils.DeleteKey(ctx, redisKey(userID)); err != nil {
		return fmt.Errorf("%w: failed to revoke stream key", ErrStreamKeyStore)
	}

	now := time.Now().UTC()
	if err := s.db.WithContext(ctx).
		Model(&database.StreamKey{}).
		Where("owner_user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{"revoked_at": now}).Error; err != nil {
		return fmt.Errorf("%w: failed to revoke stream key", ErrStreamKeyStore)
	}
	return nil
}
