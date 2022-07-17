package im

type Room struct {
	ID     int
	Master *User
	Users  []*User
}

func NewRoom(id int, master *User) *Room {
	user := master.RemovePrivacy()
	return &Room{ID: id, Master: user, Users: []*User{user}}
}

func (r *Room) AddUser(user *User) {
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
		if r.Users[i].Phone == user.Phone {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			break
		}
	}

	user.Room = nil
}
