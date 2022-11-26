package im

import (
	"encoding/json"
)

type Room struct {
	ID        int               `json:"id"`
	Master    *User             `json:"master"`
	Users     []*User           `json:"users"`
	Playlist  json.RawMessage   `json:"playlist"`
	Message   []json.RawMessage `json:"message"`
	Speed     float32           `json:"speed"`
	IsPlaying bool              `json:"is_playing"`
	Episode   int               `json:"episode"`
	Duration  int               `json:"duration"`
	SyncTime  int               `json:"sync_time"`
}

func (r *Room) AddUser(user *User) {
	for i := 0; i < len(r.Users); i++ {
		if r.Users[i].ID == user.ID {
			r.Users[i] = user
			return
		}
	}

	r.Users = append(r.Users, user)
	user.Room = r
}

func (r *Room) Dismiss() {
	for _, user := range r.Users {
		user.Room = nil
	}
	r.Master = nil
	r.Users = nil
}

func (r *Room) RemoveUser(user *User) {
	if user.Room == nil || user.Room.ID != r.ID {
		return
	}
	for i := 0; i < len(r.Users); i++ {
		if r.Users[i].ID == user.ID {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			break
		}
	}

	user.Room = nil
}
