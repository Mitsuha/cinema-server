package im

import (
	"encoding/json"
	"hourglass-socket/socket"
	"log"
	"sync"
)

type Im struct {
	ws           *socket.Service
	rooms        map[int]*Room
	users        map[string]*User
	roomCreating sync.Mutex
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
	i.ws.Listen("disconnect", i.disconnect)
	i.ws.Listen("register", i.registerUser)
	i.ws.Listen("createRoom", i.createRoom)
	i.ws.Listen("joinRoom", i.joinRoom)
	i.ws.Listen("roomInfo", i.roomInfo)
	i.ws.Listen("leaveRoom", i.leaveRoom)
	i.ws.Listen("syncPlayList", i.syncPlayList)
	i.ws.Listen("syncEpisode", i.syncEpisode)
	i.ws.Listen("syncDuration", i.syncDuration)
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

func (i *Im) disconnect(msg *socket.Message) {
	i.leaveRoom(msg)
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

	if user.Room != nil {
		i.leaveRoom(msg)
	}

	room := i.NewRoom(user)
	room.Playlist = msg.Origin

	i.rooms[room.ID], user.Room = room, room

	err := i.Reply(true, msg, room)
	if err != nil {
		log.Fatalln(err)
	}
}

func (i *Im) NewRoom(master *User) *Room {
	i.roomCreating.Lock()
	defer i.roomCreating.Unlock()

	var id = 21066
	//for true {
	//	id = rand.Intn(99999) + 10000
	//	if _, ok := i.rooms[id]; !ok {
	//		break
	//	}
	//}

	return &Room{ID: id, Master: master, Users: []*User{master}}
}

func (i *Im) leaveRoom(msg *socket.Message) {
	user, ok := msg.User().(*User)

	if ! ok || user.Room == nil || user.Room.Master == nil{
		return
	}
	if user.Room.Master.ID == user.ID {
		delete(i.rooms, user.Room.ID)
		i.BroadcastToRoom(user.Room, "dismiss", user.Room)
		user.Room.Dismiss()
		return
	}else{
		i.BroadcastToRoom(user.Room, "leaveRoom", user)
	}

	user.Room.RemoveUser(user)

	user.Room = nil
}

func (i *Im) syncPlayList(msg *socket.Message) {
	user, ok := msg.User().(*User)
	if !ok {
		return
	}
	if msg.Origin == nil {
		_ = i.Send(msg.Conn, "syncPlayList", msg.Origin)
	} else {
		user.Room.Playlist = msg.Origin
		i.BroadcastToRoom(user.Room, "syncPlayList", msg.Origin)
	}
}

func (i *Im) syncEpisode(msg *socket.Message) {
	user, ok := msg.User().(*User)
	if !ok {
		return
	}
	var data = struct {
		Index int `json:"index"`
	}{}
	if err := json.Unmarshal(msg.Origin, &data); err != nil {
		return
	}

	user.Room.Episode = data.Index
	i.BroadcastToRoom(user.Room, "syncEpisode", data)
}

func (i *Im) syncDuration(msg *socket.Message) {
	user, ok := msg.User().(*User)
	if !ok {
		return
	}
	var data = struct {
		Duration int `json:"duration"`
		Time     int `json:"time"`
	}{}

	if err := json.Unmarshal(msg.Origin, &data); err != nil {
		return
	}

	user.Room.Duration, user.Room.SyncTime = data.Duration, data.Time

	i.BroadcastToRoom(user.Room, "syncDuration", data)
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
		log.Println(err)
		_ = i.Reply(false, msg, &Response{Message: "错误的请求，也许是因为软件要更新了"})
		return
	}
	if room, ok := i.rooms[room.ID]; ok {
		room.AddUser(user)
		_ = i.Reply(true, msg, &Response{Message: "加入成功"})
		i.BroadcastToRoom(room, "joinRoom", user)
		return
	}
	_ = i.Reply(false, msg, &Response{Message: "房间不存在"})
}

func (i *Im) roomInfo(msg *socket.Message) {
	var room Room
	if err := json.Unmarshal(msg.Origin, &room); err != nil {
		log.Println(err)
		return
	}
	if room, ok := i.rooms[room.ID]; ok {
		if err := i.Reply(true, msg, room); err != nil {
			log.Println(err)
		}
		return
	}
	_ = i.Reply(false, msg, Response{Message: "房间不存在"})
}

// php -> isbn ->
// python ->
