package database

import "time"

type StreamPublishSession struct {
	ID              int64      `gorm:"column:id;primaryKey;autoIncrement"`
	StreamerUserID  int64      `gorm:"column:streamer_user_id;not null;index:idx_stream_publish_sessions_streamer_user_id"`
	MediaProviderID int64      `gorm:"column:media_provider_id;not null;index:idx_stream_publish_sessions_media_provider_id"`
	StreamKeyID     int64      `gorm:"column:stream_key_id;not null;index:idx_stream_publish_sessions_stream_key_id"`
	PlaybackID      string     `gorm:"column:playback_id;type:text;not null;uniqueIndex:idx_stream_publish_sessions_playback_id_unique"`
	Title           string     `gorm:"column:title;type:text;not null;default:''"`
	Status          string     `gorm:"column:status;type:text;not null;default:'created';index:idx_stream_publish_sessions_status"`
	PlaybackURLCDN  string     `gorm:"column:playback_url_cdn;type:text;not null;default:''"`
	CreatedAt       time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now();index:idx_stream_publish_sessions_created_at,sort:desc"`
	StartedAt       *time.Time `gorm:"column:started_at;type:timestamptz"`
	EndedAt         *time.Time `gorm:"column:ended_at;type:timestamptz"`

	StreamerUser  User          `gorm:"foreignKey:StreamerUserID;references:ID"`
	MediaProvider MediaProvider `gorm:"foreignKey:MediaProviderID;references:ID"`
	StreamKey     StreamKey     `gorm:"foreignKey:StreamKeyID;references:ID"`
	Renditions    []TranscodeRendition `gorm:"foreignKey:PublishSessionID;references:ID"`
}
