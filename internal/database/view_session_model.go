package database

import "time"

type ViewSession struct {
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PublishSessionID int64      `gorm:"column:publish_session_id;not null;index:idx_view_sessions_publish_session_id"`
	ViewerUserID     *int64     `gorm:"column:viewer_user_id;index:idx_view_sessions_viewer_user_id"`
	ViewerRef        string     `gorm:"column:viewer_ref;type:text;not null;default:''"`
	ClientType       string     `gorm:"column:client_type;type:text;not null;default:'web'"`
	JoinedAt         time.Time  `gorm:"column:joined_at;type:timestamptz;not null;default:now();index:idx_view_sessions_joined_at,sort:desc"`
	LeftAt           *time.Time `gorm:"column:left_at;type:timestamptz"`
	LastSeenAt       time.Time  `gorm:"column:last_seen_at;type:timestamptz;not null;default:now()"`

	PublishSession StreamPublishSession `gorm:"foreignKey:PublishSessionID;references:ID"`
	ViewerUser     *User                `gorm:"foreignKey:ViewerUserID;references:ID"`
}
