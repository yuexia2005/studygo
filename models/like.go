package models

import "time"

type Like struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index;not null"`
	VideoID   uint `gorm:"index;not null;constraint:OnDelete:CASCADE"`
	CreatedAt time.Time
}
