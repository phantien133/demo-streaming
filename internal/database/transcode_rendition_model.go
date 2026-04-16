package database

import "time"

type TranscodeRendition struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PublishSessionID int64    `gorm:"column:publish_session_id;not null;index:idx_transcode_renditions_publish_session_id;uniqueIndex:idx_transcode_renditions_unique"`
	PlaybackID      string    `gorm:"column:playback_id;type:text;not null;index:idx_transcode_renditions_playback_id"`
	RenditionName   string    `gorm:"column:rendition_name;type:text;not null;uniqueIndex:idx_transcode_renditions_unique"`
	PlaylistPath    string    `gorm:"column:playlist_path;type:text;not null"`
	Status          string    `gorm:"column:status;type:text;not null;default:'ready'"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now()"`

	PublishSession StreamPublishSession `gorm:"foreignKey:PublishSessionID;references:ID"`
}
