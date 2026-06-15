package models

import "time"

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Username  string `gorm:"unique;size:50;not null"`
	Password  string `gorm:"size:100;not null"`
	CreatedAt time.Time
}
