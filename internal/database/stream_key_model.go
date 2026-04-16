package database

import "time"

type StreamKey struct {
	ID              int64      `gorm:"column:id;primaryKey;autoIncrement"`
	OwnerUserID     int64      `gorm:"column:owner_user_id;not null;index:idx_stream_keys_owner_user_id"`
	StreamKeySecret string     `gorm:"column:stream_key_secret;type:text;not null;uniqueIndex"`
	MediaProviderID int64      `gorm:"column:media_provider_id;not null"`
	Label           string     `gorm:"column:label;type:text;not null;default:''"`
	CreatedAt       time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
	RevokedAt       *time.Time `gorm:"column:revoked_at;type:timestamptz"`

	OwnerUser     User          `gorm:"foreignKey:OwnerUserID;references:ID"`
	MediaProvider MediaProvider `gorm:"foreignKey:MediaProviderID;references:ID"`
}
