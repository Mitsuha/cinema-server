package model

import (
	"hourglass-socket/socket"
)

type User struct {
	ID     string          `json:"id"`
	Avatar string          `json:"avatar"`
	Name   string          `json:"name"`
	Room   *Room           `json:"-"`
	Conn   *socket.Connect `json:"-"`
}

func (u *User) Association(connect *socket.Connect) {
	u.Conn = connect
	connect.Attach = u
}
