package model

import "time"

type User struct {
	ID        int        `json:"id" gorm:"id"`
	AliID     string     `json:"ali_id" gorm:"ali_id"`
	Avatar    string     `json:"avatar" gorm:"avatar"`
	Name      string     `json:"name" gorm:"name"`
	Phone     string     `json:"phone" gorm:"phone"`
	CreatedAt *time.Time `json:"created_at" gorm:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" gorm:"updated_at"`
}
