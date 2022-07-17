package im

import "hourglass-socket/socket"

type User struct {
	ID     string          `json:"id"`
	Avatar string          `json:"avatar"`
	Name   string          `json:"name"`
	Phone  string          `json:"phone"`
	Room   *Room           `json:"-"`
	Conn   *socket.Connect `json:"-"`
}

func (u *User) RemovePrivacy() *User {
	return &User{
		ID:     u.ID,
		Avatar: u.Avatar,
		Name:   u.Name,
	}
}

func (u *User) association(connect *socket.Connect) {
	u.Conn = connect
	connect.User = u
}
