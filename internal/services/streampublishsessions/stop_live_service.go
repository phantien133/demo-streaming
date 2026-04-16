package streampublishsessions

import (
	"context"
	"errors"
	"time"

	"demo-streaming/internal/database"
	"gorm.io/gorm"
)

type GormStopLiveService struct {
	db *gorm.DB
}

func NewGormStopLiveService(db *gorm.DB) *GormStopLiveService {
	return &GormStopLiveService{db: db}
}

func (s *GormStopLiveService) Execute(ctx context.Context, input StopLiveInput) (StopLiveOutput, error) {
	if input.StreamerUserID <= 0 {
		return StopLiveOutput{}, ErrPublishSessionForbidden
	}

	endedAt := time.Now().UTC()
	db := s.db.WithContext(ctx)

	// Stop a specific session.
	if input.SessionID != nil {
		var session database.StreamPublishSession
		if err := db.First(&session, "id = ?", *input.SessionID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return StopLiveOutput{}, ErrPublishSessionNotFound
			}
			return StopLiveOutput{}, err
		}
		if session.StreamerUserID != input.StreamerUserID {
			return StopLiveOutput{}, ErrPublishSessionForbidden
		}
		if session.Status != "live" {
			return StopLiveOutput{}, ErrPublishSessionBadState
		}

		if err := db.Model(&database.StreamPublishSession{}).
			Where("id = ?", session.ID).
			Updates(map[string]any{
				"status":   "ended",
				"ended_at": endedAt,
			}).Error; err != nil {
			return StopLiveOutput{}, err
		}
		if err := db.Model(&database.ViewSession{}).
			Where("publish_session_id = ? AND left_at IS NULL", session.ID).
			Updates(map[string]any{
				"left_at":      endedAt,
				"last_seen_at": endedAt,
			}).Error; err != nil {
			return StopLiveOutput{}, err
		}

		return StopLiveOutput{
			StoppedSessionIDs: []int64{session.ID},
			EndedAt:           endedAt,
		}, nil
	}

	// Stop ALL live sessions for this streamer.
	var liveSessionIDs []int64
	if err := db.Model(&database.StreamPublishSession{}).
		Where("streamer_user_id = ? AND status = ?", input.StreamerUserID, "live").
		Pluck("id", &liveSessionIDs).Error; err != nil {
		return StopLiveOutput{}, err
	}

	if len(liveSessionIDs) == 0 {
		return StopLiveOutput{
			StoppedSessionIDs: nil,
			EndedAt:           endedAt,
		}, nil
	}

	if err := db.Model(&database.StreamPublishSession{}).
		Where("id IN ?", liveSessionIDs).
		Updates(map[string]any{
			"status":   "ended",
			"ended_at": endedAt,
		}).Error; err != nil {
		return StopLiveOutput{}, err
	}
	if err := db.Model(&database.ViewSession{}).
		Where("publish_session_id IN ? AND left_at IS NULL", liveSessionIDs).
		Updates(map[string]any{
			"left_at":      endedAt,
			"last_seen_at": endedAt,
		}).Error; err != nil {
		return StopLiveOutput{}, err
	}

	return StopLiveOutput{
		StoppedSessionIDs: liveSessionIDs,
		EndedAt:           endedAt,
	}, nil
}

