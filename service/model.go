package service

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"hourglass-socket/model"
)

type User struct {
	model.User
	Conn *websocket.Conn `json:"-" gorm:"-"`
}

type Room struct {
	ModelID   int             `json:"-"`
	Code      int             `json:"id"`
	Master    *User           `json:"master"`
	Users     []*User         `json:"users"`
	Playlist  json.RawMessage `json:"playlist"`
	Message   []model.Message `json:"message"`
	Speed     float32         `json:"speed"`
	IsPlaying bool            `json:"is_playing"`
	Episode   int             `json:"episode"`
	Duration  int             `json:"duration"`
	OnDismiss chan bool       `json:"-"`
}

func (r *Room) ToModel() *model.Room {
	var uid = make([]int, len(r.Users))

	for i, user := range r.Users {
		uid[i] = user.ID
	}

	return &model.Room{
		UserID:   r.Master.ID,
		Audience: uid,
	}
}
