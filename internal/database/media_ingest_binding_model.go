package database

import "time"

type MediaIngestBinding struct {
	ID                 int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PublishSessionID   int64      `gorm:"column:publish_session_id;not null;uniqueIndex"`
	ProviderPublishID  *string    `gorm:"column:provider_publish_id;type:text;index:idx_media_ingest_bindings_provider_publish_id"`
	ProviderVHost      *string    `gorm:"column:provider_vhost;type:text"`
	ProviderApp        *string    `gorm:"column:provider_app;type:text"`
	IngestQueryParam   *string    `gorm:"column:ingest_query_param;type:text"`
	RecordLocalURI     *string    `gorm:"column:record_local_uri;type:text"`
	LastCallbackAction *string    `gorm:"column:last_callback_action;type:text"`
	LastCallbackAt     *time.Time `gorm:"column:last_callback_at;type:timestamptz"`
	ProviderContext    []byte     `gorm:"column:provider_context;type:jsonb;not null;default:'{}'"`
	CreatedAt          time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;type:timestamptz;not null;default:now()"`

	PublishSession StreamPublishSession `gorm:"foreignKey:PublishSessionID;references:ID"`
}
