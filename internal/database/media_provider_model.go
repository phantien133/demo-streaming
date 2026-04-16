package database

import "time"

type MediaProvider struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Code        string    `gorm:"column:code;type:text;not null;uniqueIndex"`
	DisplayName string    `gorm:"column:display_name;type:text;not null;default:''"`
	APIBaseURL  *string   `gorm:"column:api_base_url;type:text"`
	Config      []byte    `gorm:"column:config;type:jsonb;not null;default:'{}'"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
}
