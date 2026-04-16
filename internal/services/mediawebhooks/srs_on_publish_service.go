package mediawebhooks

import (
	"context"
	"errors"
	"strings"
	"time"

	"demo-streaming/internal/database"
	transcodeservice "demo-streaming/internal/services/transcode"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadRequest   = errors.New("bad request")
)

type SRSOnPublishInput struct {
	// Auth
	WebhookSecretHeader string
	ExpectedSecret      string

	// SRS payload
	Stream string
	Param  string
}

type SRSOnPublishOutput struct {
	SessionID int64
	Status    string
}

type SRSOnPublishService interface {
	Execute(ctx context.Context, input SRSOnPublishInput) (SRSOnPublishOutput, error)
}

type GormSRSOnPublishService struct {
	db               *gorm.DB
	jobQueueEnqueuer transcodeservice.PublishJobEnqueuer
}

func NewGormSRSOnPublishService(db *gorm.DB, enqueuer transcodeservice.PublishJobEnqueuer) *GormSRSOnPublishService {
	return &GormSRSOnPublishService{
		db:               db,
		jobQueueEnqueuer: enqueuer,
	}
}

func (s *GormSRSOnPublishService) Execute(ctx context.Context, input SRSOnPublishInput) (SRSOnPublishOutput, error) {
	if strings.TrimSpace(input.ExpectedSecret) != "" && input.WebhookSecretHeader != input.ExpectedSecret {
		return SRSOnPublishOutput{}, ErrUnauthorized
	}
	streamName := strings.TrimSpace(input.Stream)
	if streamName == "" {
		return SRSOnPublishOutput{}, ErrBadRequest
	}

	// streamName is the RTMP app stream name (ingest key). Resolve session by stream_keys first; fall back to
	// playback_id for older rows where stream name matched the public id.
	activeStatuses := []string{"created", "live"}

	var session database.StreamPublishSession
	var streamKey database.StreamKey
	keyErr := s.db.WithContext(ctx).
		Where("stream_key_secret = ? AND revoked_at IS NULL", streamName).
		First(&streamKey).Error
	if keyErr == nil {
		sessErr := s.db.WithContext(ctx).
			Where("stream_key_id = ? AND status IN ?", streamKey.ID, activeStatuses).
			Order("id DESC").
			First(&session).Error
		if sessErr != nil && !errors.Is(sessErr, gorm.ErrRecordNotFound) {
			return SRSOnPublishOutput{}, sessErr
		}
	} else if !errors.Is(keyErr, gorm.ErrRecordNotFound) {
		return SRSOnPublishOutput{}, keyErr
	}

	if session.ID == 0 {
		if err := s.db.WithContext(ctx).
			Where("playback_id = ? AND status IN ?", streamName, activeStatuses).
			Order("id DESC").
			First(&session).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return SRSOnPublishOutput{}, ErrForbidden
			}
			return SRSOnPublishOutput{}, err
		}
	}

	var sk database.StreamKey
	if err := s.db.WithContext(ctx).First(&sk, session.StreamKeyID).Error; err != nil {
		return SRSOnPublishOutput{}, err
	}

	// Upsert media_ingest_bindings for traceability.
	now := time.Now().UTC()
	action := "on_publish"
	param := strings.TrimSpace(input.Param)
	providerPublishID := streamName
	binding := database.MediaIngestBinding{
		PublishSessionID:   session.ID,
		ProviderPublishID:  &providerPublishID,
		IngestQueryParam:   &param,
		LastCallbackAction: &action,
		LastCallbackAt:     &now,
	}
	if err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "publish_session_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"provider_publish_id", "ingest_query_param", "last_callback_action", "last_callback_at", "updated_at"}),
		}).
		Create(&binding).Error; err != nil {
		return SRSOnPublishOutput{}, err
	}

	if s.jobQueueEnqueuer != nil {
		// Best-effort enqueue: publish availability should not depend on worker queue.
		_ = s.jobQueueEnqueuer.EnqueuePublishJob(ctx, transcodeservice.PublishJob{
			SessionID:       session.ID,
			PlaybackID:      session.PlaybackID,
			StreamKeySecret: strings.TrimSpace(sk.StreamKeySecret),
		})
	}

	return SRSOnPublishOutput{
		SessionID: session.ID,
		Status:    session.Status,
	}, nil
}
