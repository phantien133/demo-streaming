package streampublishsessions

import (
	"context"
	"errors"
	"time"

	"demo-streaming/internal/database"
	"gorm.io/gorm"
)

type GormStartService struct {
	db *gorm.DB
}

func NewGormStartService(db *gorm.DB) *GormStartService {
	return &GormStartService{db: db}
}

func (s *GormStartService) Execute(ctx context.Context, input StartInput) (StartOutput, error) {
	var session database.StreamPublishSession
	if err := s.db.WithContext(ctx).First(&session, "id = ?", input.SessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return StartOutput{}, ErrPublishSessionNotFound
		}
		return StartOutput{}, err
	}

	if session.StreamerUserID != input.StreamerUserID {
		return StartOutput{}, ErrPublishSessionForbidden
	}
	if session.Status != "created" {
		return StartOutput{}, ErrPublishSessionBadState
	}

	startedAt := time.Now().UTC()
	if err := s.db.WithContext(ctx).
		Model(&database.StreamPublishSession{}).
		Where("id = ?", session.ID).
		Updates(map[string]any{
			"status":     "live",
			"started_at": startedAt,
		}).Error; err != nil {
		return StartOutput{}, err
	}

	return StartOutput{
		SessionID: session.ID,
		Status:    "live",
		StartedAt: startedAt,
	}, nil
}
