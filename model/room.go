package model

import "time"

type Room struct {
	ID        int        `json:"id" gorm:"id"`
	UserID    int        `json:"master" gorm:"master"`
	Audience  []int      `json:"audience" gorm:"audience;serializer:json"`
	Duration  int        `json:"duration" gorm:"duration"`
	DismissAt *time.Time `json:"dismiss_at" gorm:"dismiss_at"`
	CreatedAt *time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"updated_at"`
}
