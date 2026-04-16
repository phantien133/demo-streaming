package database

import "time"

type User struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Email       string    `gorm:"column:email;type:text;not null;uniqueIndex"`
	DisplayName string    `gorm:"column:display_name;type:text;not null;default:''"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
}
