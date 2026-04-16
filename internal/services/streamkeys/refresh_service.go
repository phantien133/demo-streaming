package streamkeys

import (
	"context"
	"errors"
	"fmt"
	"time"

	"demo-streaming/internal/database"
	redisutil "demo-streaming/internal/utils/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RefreshStreamKeyService struct {
	db        *gorm.DB
	redisUtils *redisutil.RedisUtils
}

func NewRefreshStreamKeyService(db *gorm.DB, redisClient *redis.Client) *RefreshStreamKeyService {
	return &RefreshStreamKeyService{
		db:        db,
		redisUtils: redisutil.NewRedisUtils(redisClient),
	}
}

func (s *RefreshStreamKeyService) Execute(ctx context.Context, userID int64) (string, int64, error) {
	if s.db == nil {
		return "", 0, fmt.Errorf("%w: db not configured", ErrStreamKeyStore)
	}

	// Require an existing active key to refresh.
	var existing database.StreamKey
	if err := s.db.WithContext(ctx).
		Where("owner_user_id = ? AND revoked_at IS NULL", userID).
		Order("id DESC").
		First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", 0, ErrStreamKeyNotFound
		}
		return "", 0, fmt.Errorf("%w: failed to query stream key", ErrStreamKeyStore)
	}

	streamKey, err := GenerateStreamKey()
	if err != nil {
		return "", 0, err
	}

	// Revoke any existing active key.
	now := time.Now().UTC()
	if err := s.db.WithContext(ctx).
		Model(&database.StreamKey{}).
		Where("owner_user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{"revoked_at": now}).Error; err != nil {
		return "", 0, fmt.Errorf("%w: failed to revoke existing stream key", ErrStreamKeyStore)
	}

	var provider database.MediaProvider
	if err := s.db.WithContext(ctx).Where("code = ?", "srs").First(&provider).Error; err != nil {
		return "", 0, fmt.Errorf("%w: media provider not configured", ErrStreamKeyStore)
	}
	if err := s.db.WithContext(ctx).Create(&database.StreamKey{
		OwnerUserID:     userID,
		StreamKeySecret: streamKey,
		MediaProviderID: provider.ID,
	}).Error; err != nil {
		return "", 0, fmt.Errorf("%w: failed to persist stream key", ErrStreamKeyStore)
	}

	key := redisKey(userID)
	if err := s.redisUtils.SetString(ctx, key, streamKey, defaultStreamKeyTTL); err != nil {
		return "", 0, fmt.Errorf("%w: failed to store stream key", ErrStreamKeyStore)
	}

	return streamKey, int64(defaultStreamKeyTTL.Seconds()), nil
}
