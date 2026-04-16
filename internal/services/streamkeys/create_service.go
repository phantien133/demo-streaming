package streamkeys

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"demo-streaming/internal/database"
	redisutil "demo-streaming/internal/utils/redis"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CreateStreamKeyService struct {
	db        *gorm.DB
	client    *redis.Client
	redisUtils *redisutil.RedisUtils
}

func NewCreateStreamKeyService(db *gorm.DB, redisClient *redis.Client) *CreateStreamKeyService {
	return &CreateStreamKeyService{
		db:        db,
		client:    redisClient,
		redisUtils: redisutil.NewRedisUtils(redisClient),
	}
}

func (s *CreateStreamKeyService) Execute(ctx context.Context, userID int64, expiresInSeconds *int64) (string, int64, error) {
	if s.db == nil {
		return "", 0, fmt.Errorf("%w: db not configured", ErrStreamKeyStore)
	}

	ttl := defaultStreamKeyTTL
	if expiresInSeconds != nil && *expiresInSeconds > 0 {
		ttl = time.Duration(*expiresInSeconds) * time.Second
	}

	// DB is the source of truth for stream keys, so SRS webhooks can validate.
	var existing database.StreamKey
	if err := s.db.WithContext(ctx).
		Where("owner_user_id = ? AND revoked_at IS NULL", userID).
		Order("id DESC").
		First(&existing).Error; err == nil {
		streamKey := strings.TrimSpace(existing.StreamKeySecret)
		if streamKey == "" {
			return "", 0, fmt.Errorf("%w: existing stream key empty", ErrStreamKeyStore)
		}
		if err := s.redisUtils.SetString(ctx, redisKey(userID), streamKey, ttl); err != nil {
			return "", 0, fmt.Errorf("%w: failed to cache stream key", ErrStreamKeyStore)
		}
		return streamKey, int64(ttl.Seconds()), nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", 0, fmt.Errorf("%w: failed to query stream key", ErrStreamKeyStore)
	}

	// Create a brand new stream key and persist it.
	streamKey, err := GenerateStreamKey()
	if err != nil {
		return "", 0, err
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

	if err := s.redisUtils.SetString(ctx, redisKey(userID), streamKey, ttl); err != nil {
		return "", 0, fmt.Errorf("%w: failed to cache stream key", ErrStreamKeyStore)
	}
	return streamKey, int64(ttl.Seconds()), nil
}
