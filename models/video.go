package models

import "time"

type Video struct {
	ID          uint   `gorm:"primaryKey"`
	Title       string `gorm:"size:200;not null"`
	Description string `gorm:"size:500"`
	FilePath    string `gorm:"size:255;not null"`
	CoverURL    string `gorm:"size:255"`
	UserID      uint   `gorm:"index;not null"`
	ViewCount   int    `gorm:"default:0"`
	LikeCount   int    `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
