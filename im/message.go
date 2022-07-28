package im

import (
	"time"
)

type ChatMessage struct {
	Id         uint
	SenderId   uint
	ReceiverId uint
	Type       int
	Content    string
	Annex      string
	UUID       string
	Received   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Response struct {
	Success bool `json:"success"`
	Message string `json:"message"`
}
