package models

import "time"

type Comment struct {
	ID        uint   `gorm:"primaryKey"`
	Content   string `gorm:"type:text;not null"`
	UserID    uint   `gorm:"index;not null"`
	VideoID   uint   `gorm:"index;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
