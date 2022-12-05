package model

import "time"

type Message struct {
	ID        int       `json:"id"`
	RoomID    int       `json:"roomID"`
	UserID    int       `json:"userID"`
	Content   string    `json:"content"`
	UUID      string    `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
