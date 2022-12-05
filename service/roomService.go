package service

import (
	"errors"
	"math/rand"
	"time"
)

type RoomService struct {
	rooms  map[int]*Room
	idChan chan int
}

var roomService = makeRoomService()

func makeRoomService() *RoomService {
	service := RoomService{
		rooms:  map[int]*Room{},
		idChan: make(chan int),
	}
	service.Boot()

	return &service
}

func (r *RoomService) Boot() {
	go func() {
		var id int
		for true {
			rand.Seed(time.Now().UnixNano())
			id = rand.Intn(99999) + 10000
			if _, ok := r.rooms[id]; !ok {
				r.idChan <- id
			}
		}
	}()
}

func (r *RoomService) Create(user *User) *Room {
	var room = Room{
		Code:      <-r.idChan,
		Master:    user,
		Users:     []*User{user},
		Speed:     1,
		IsPlaying: true,
		Playlist:  nil,
		Message:   nil,
		Episode:   0,
		Duration:  1,
		OnDismiss: make(chan bool),
	}

	r.rooms[room.Code] = &room

	return &room
}

func (r *RoomService) Find(code int) *Room {
	room, _ := r.rooms[code]
	return room
}

func (r *RoomService) Exist(id int) bool {
	_, ok := r.rooms[id]
	return ok
}

func (r *RoomService) Dismiss(room *Room) {
	if room, ok := r.rooms[room.Code]; ok {
		close(room.OnDismiss)
		delete(r.rooms, room.Code)
	}
}

func (r *RoomService) RemoteUser(room *Room, user *User) {
	for i, u := range room.Users {
		if u.AliID == user.AliID {
			room.Users = append(room.Users[:i], room.Users[i+1:]...)
		}
	}
}

func (r *RoomService) JoinRoom(room *Room, user *User) (*Room, error) {
	if room, ok := r.rooms[room.Code]; ok {
		for i, u := range room.Users {
			if u.AliID == user.AliID {
				room.Users[i] = user
				return room, nil
			}
		}

		room.Users = append(room.Users, user)
		return room, nil
	} else {
		return nil, errors.New("room not exist")
	}
}
