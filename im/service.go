package im

import (
	"encoding/json"
	"hourglass-socket/socket"
	"log"
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

func (i *Im) registerUser(message *socket.Message) {
	user := User{}

	if err := json.Unmarshal(message.Origin, &user); err != nil {
		log.Fatalln(err)
		return
	}

	if u, ok := i.users[user.ID]; ok {
		u.association(message.Conn)
	} else {
		user.association(message.Conn)
		i.users[user.ID] = &user
	}
}

func (i *Im) createRoom(msg *socket.Message) {
	user, ok := msg.User().(*User)

	if user == nil || !ok {
		err := i.Reply(true, msg, &Response{Message: "未登录，试试重启？"})
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	//if user.Room != nil {
	//	err := i.Reply(msg, &Response{Message: "你已经在一个房间里，请退出后再试"})
	//	if err != nil {
	//		log.Fatalln(err)
	//	}
	//	return
	//}

	room := i.NewRoom(user)

	i.rooms[i.incrementRoomID], user.Room = room, room

	err := i.Reply(true, msg, room)
	if err != nil {
		log.Fatalln(err)
	}
}

func (i *Im) NewRoom(master *User) *Room {
	i.roomCreating.Lock()
	defer func() {
		i.incrementRoomID++
		i.roomCreating.Unlock()
	}()

	user := master.RemovePrivacy()

	return &Room{ID: i.incrementRoomID, Master: user, Users: []*User{user}}
}

func (i *Im) leaveRoom(msg *socket.Message) {
	user, ok := msg.User().(*User)
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

func (i *Im) broadcast(msg *socket.Message) {
	user, ok := msg.User().(*User)
	if !ok {
		return
	}

	i.BroadcastToRoom(user.Room, msg.Event, msg.Origin)
}

func (i *Im) joinRoom(msg *socket.Message) {
	user, ok := msg.User().(*User)
	if !ok {
		_ = i.Reply(false, msg, &Response{Message: "未登录，试试重启？"})
		return
	}
	room := Room{}
	err := json.Unmarshal(msg.Origin, &room)
	if err != nil {
		print(err.Error())
		_ = i.Reply(false, msg, &Response{Message: "错误的请求，也许是因为软件要更新了"})
		return
	}
	if room, ok := i.rooms[room.ID]; ok {
		room.AddUser(user)
		i.BroadcastToRoom(room, "joinRoom", user.RemovePrivacy())
		return
	}
	_ = i.Reply(false, msg, &Response{Message: "房间不存在"})
}

func (i *Im) Reply(success bool, msg *socket.Message, message interface{}) error {
	return i.ws.Emit(msg.Conn, &socket.Message{
		ID:      msg.ID,
		Success: success,
		Event:   "reply",
		Payload: message,
	})
}

func (i *Im) Send(conn *socket.Connect, event string, message interface{}) error {
	return i.ws.Emit(conn, &socket.Message{
		Event:   event,
		Payload: message,
	})

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
