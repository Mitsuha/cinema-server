package im

import "hourglass-socket/socket"

type User struct {
	ID     string          `json:"id"`
	Avatar string          `json:"avatar"`
	Name   string          `json:"name"`
	Room   *Room           `json:"-"`
	Conn   *socket.Connect `json:"-"`
}

func (u *User) association(connect *socket.Connect) {
	u.Conn = connect
	connect.User = u
}
