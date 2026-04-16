package transcode

import (
	"context"
	"time"

	"demo-streaming/internal/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormMetadataStore struct {
	db *gorm.DB
}

func NewGormMetadataStore(db *gorm.DB) *GormMetadataStore {
	return &GormMetadataStore{db: db}
}

func (s *GormMetadataStore) Record(ctx context.Context, job PublishJob, result TranscodeResult) error {
	if len(result.Renditions) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for _, output := range result.Renditions {
		row := database.TranscodeRendition{
			PublishSessionID: job.SessionID,
			PlaybackID:       job.PlaybackID,
			RenditionName:    output.Name,
			PlaylistPath:     output.PlaylistPath,
			Status:           output.Status,
			UpdatedAt:        now,
		}
		if err := s.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "publish_session_id"},
					{Name: "rendition_name"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"playback_id",
					"playlist_path",
					"status",
					"updated_at",
				}),
			}).
			Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}
