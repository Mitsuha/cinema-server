package im

import (
	"encoding/json"
	"hourglass-socket/socket"
	"sync"
)

type Im struct {
	ws              *socket.Service
	rooms           map[int]*Room
	users           map[string]*User
	incrementRoomID int
	roomCreating    sync.Mutex
}

func New(ws *socket.Service) *Im {
	return &Im{
		ws:    ws,
		rooms: make(map[int]*Room),
		users: make(map[string]*User),
	}
}

func (i *Im) Init() {
	i.RegisterEvent()
}

func (i *Im) RegisterEvent() {
	i.ws.Listen("register", i.registerUser)
	i.ws.Listen("createRoom", i.createRoom)
	i.ws.Listen("joinRoom", i.joinRoom)
	i.ws.Listen("leaveRoom", i.leaveRoom)
	i.ws.Listen("broadcast", i.broadcast)
}

func (i *Im) registerUser(conn *socket.Connect, message *socket.Message) {
	user := User{}

	if err := json.Unmarshal(message.Origin, &user); err != nil {
		return
	}

	if u, ok := i.users[user.Phone]; ok {
		u.association(conn)
	} else {
		user.association(conn)
		i.users[user.Phone] = &user
	}
}

func (i *Im) createRoom(conn *socket.Connect, _ *socket.Message) {
	user, ok := conn.User.(*User)
	if !ok {
		return
	}
	if user == nil {
		_ = i.Reply(conn, &Response{Message: "未登录，试试重启？"})
		return
	}

	if user.Room != nil {
		_ = i.Reply(conn, &Response{Message: "你已经在一个房间里，请退出后再试"})
		return
	}

	i.roomCreating.Lock()
	room := NewRoom(i.incrementRoomID, user)
	i.rooms[i.incrementRoomID] = room
	user.Room = room
	i.incrementRoomID++
	i.roomCreating.Unlock()

	_ = i.Reply(conn, room)
}

func (i *Im) leaveRoom(conn *socket.Connect, _ *socket.Message) {
	user, ok := conn.User.(*User)
	if !ok {
		return
	}
	if user.Room == nil {
		return
	}
	if user.Room.Master.ID == user.ID {
		delete(i.rooms, user.Room.ID)
		i.BroadcastToRoom(user.Room, "dismiss", user.Room)
		user.Room.Dismiss()
		return
	}

	user.Room.RemoveUser(user)

	i.BroadcastToRoom(user.Room, "leave", user.RemovePrivacy())
	user.Room = nil
}

func (i *Im) broadcast(conn *socket.Connect, message *socket.Message) {
	user, ok := conn.User.(*User)
	if !ok {
		return
	}

	i.BroadcastToRoom(user.Room, message.Event, message.Origin)
}

func (i *Im) joinRoom(conn *socket.Connect, message *socket.Message) {
	user, ok := conn.User.(*User)
	if !ok {
		_ = i.Reply(conn, &Response{Message: "未登录，试试重启？"})
		return
	}
	room := Room{}
	err := json.Unmarshal(message.Origin, &room)
	if err != nil {
		print(err.Error())
		_ = i.Reply(conn, &Response{Message: "错误的请求，也许是因为软件要更新了"})
		return
	}
	if room, ok := i.rooms[room.ID]; ok {
		room.AddUser(user)
		i.BroadcastToRoom(room, "joinRoom", user.RemovePrivacy())
		return
	}
	_ = i.Reply(conn, &Response{Message: "房间不存在"})
}

func (i *Im) Reply(conn *socket.Connect, response interface{}) error {
	return i.Send(conn, "reply", response)
}

func (i *Im) Send(conn *socket.Connect, event string, message interface{}) error {
	return i.ws.Send(conn, event, message)
}

func (i *Im) BroadcastToRoom(room *Room, event string, message interface{}) []error {
	errs := make([]error, 0)
	for _, user := range room.Users {
		err := i.Send(user.Conn, event, message)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
