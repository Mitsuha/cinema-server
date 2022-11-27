package im

import (
	"encoding/json"
	"hourglass-socket/distribution"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Im struct {
	distribution *distribution.Distribution
	rooms        map[int]*Room
	users        map[string]*User
	roomCreating sync.Mutex
}

func New(distribution *distribution.Distribution) *Im {
	return &Im{
		distribution: distribution,
		rooms:        make(map[int]*Room),
		users:        make(map[string]*User),
	}
}

func (i *Im) Init() {
	i.RegisterEvent()
}

func (i *Im) RegisterEvent() {
	i.distribution.Register("disconnect", distribution.Listener{
		Action: i.disconnect,
	})
	i.distribution.Register("register", distribution.Listener{
		Action: i.registerUser,
	})
	i.distribution.Register("createRoom", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth},
		Action:      i.createRoom,
	})
	i.distribution.Register("joinRoom", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth},
		Action:      i.joinRoom,
	})
	i.distribution.Register("roomInfo", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth},
		Action:      i.roomInfo,
	})
	i.distribution.Register("leaveRoom", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth, i.HasRoom},
		Action:      i.leaveRoom,
	})
	i.distribution.Register("syncPlayList", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth, i.HasRoom},
		Action:      i.syncPlayList,
	})
	i.distribution.Register("syncEpisode", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth, i.HasRoom},
		Action:      i.syncEpisode,
	})
	i.distribution.Register("syncDuration", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth, i.HasRoom},
		Action:      i.syncDuration,
	})
	i.distribution.Register("syncSpeed", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth, i.HasRoom},
		Action:      i.syncSpeed,
	})
	i.distribution.Register("syncPlayingStatus", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth, i.HasRoom},
		Action:      i.syncPlayingStatus,
	})
	i.distribution.Register("chat", distribution.Listener{
		Middlewares: []distribution.Middleware{i.Auth, i.HasRoom},
		Action:      i.chat,
	})
}

func (i *Im) registerUser(message *distribution.Message) {
	user := User{}

	if err := json.Unmarshal(message.Origin, &user); err != nil {
		log.Println(err)
		return
	}

	if u, ok := i.users[user.ID]; ok {
		u.association(message.Conn)
	} else {
		user.association(message.Conn)
		i.users[user.ID] = &user
	}
}

func (i *Im) disconnect(msg *distribution.Message) {
	if user, _ := msg.User().(*User); user != nil {
		i.leaveRoom(msg)
	}
}

func (i *Im) createRoom(msg *distribution.Message) {
	user, _ := msg.User().(*User)

	if user.Room != nil {
		i.leaveRoom(msg)
	}

	room := i.NewRoom(user)
	room.Playlist = msg.Origin

	i.rooms[room.ID], user.Room = room, room

	err := i.distribution.Reply(true, msg, room)
	if err != nil {
		log.Println(err)
	}
}

func (i *Im) NewRoom(master *User) *Room {
	i.roomCreating.Lock()
	defer i.roomCreating.Unlock()

	var id int
	for true {
		rand.Seed(time.Now().UnixNano())
		id = rand.Intn(99999) + 10000
		if _, ok := i.rooms[id]; !ok {
			break
		}
	}

	return &Room{
		ID:        id,
		Master:    master,
		Users:     []*User{master},
		Speed:     1,
		IsPlaying: true,
	}
}

func (i *Im) leaveRoom(msg *distribution.Message) {
	user, _ := msg.User().(*User)

	if user == nil || user.Room == nil || user.Room.Master == nil {
		return
	}

	// dismiss room
	if user.Room.Master.ID == user.ID {
		delete(i.rooms, user.Room.ID)
		i.BroadcastToRoom(user.Room, "dismiss", user.Room)
		user.Room.Dismiss()
		return
	} else {
		i.BroadcastToRoom(user.Room, "leaveRoom", user)

		user.Room.RemoveUser(user)
		user.Room = nil
	}
}

func (i *Im) syncPlayList(msg *distribution.Message) {
	user, _ := msg.User().(*User)

	data := struct {
		Playlist json.RawMessage `json:"playlist"`
	}{}

	if err := json.Unmarshal(msg.Origin, &data); err == nil {
		user.Room.Playlist = data.Playlist
		i.BroadcastToRoom(user.Room, "syncPlayList", msg.Origin)
	}
}

func (i *Im) syncEpisode(msg *distribution.Message) {
	user, _ := msg.User().(*User)

	var data = struct {
		Index    int               `json:"index"`
		PlayInfo []json.RawMessage `json:"playInfo"`
	}{}

	if err := json.Unmarshal(msg.Origin, &data); err == nil && len(data.PlayInfo) != 0 {
		var cursor = 0
		for _, u := range user.Room.Users {
			if u.ID != user.Room.Master.ID {
				_ = i.distribution.Send(u.Conn, "syncEpisode", map[string]interface{}{
					"index":    data.Index,
					"playInfo": data.PlayInfo[cursor%len(data.PlayInfo)],
				})
				cursor++
			}
		}

		user.Room.Episode, user.Room.Duration, user.Room.Speed = data.Index, 0, 1
	} else {
		log.Println(err)
	}
}

func (i *Im) syncDuration(msg *distribution.Message) {
	user, _ := msg.User().(*User)

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

func (i *Im) syncSpeed(msg *distribution.Message) {
	user, _ := msg.User().(*User)

	data := struct {
		Speed float32 `json:"speed"`
	}{}

	if json.Unmarshal(msg.Origin, &data) == nil {
		user.Room.Speed = data.Speed

		i.BroadcastToRoom(user.Room, "syncSpeed", msg.Origin)
	}

}

func (i *Im) syncPlayingStatus(msg *distribution.Message) {
	user, _ := msg.User().(*User)

	var data = struct {
		Playing bool `json:"playing"`
	}{}

	if err := json.Unmarshal(msg.Origin, &data); err != nil {
		return
	}

	user.Room.IsPlaying = data.Playing

	i.BroadcastToRoom(user.Room, "syncPlayingStatus", data)
}

func (i *Im) chat(msg *distribution.Message) {
	user, _ := msg.User().(*User)

	if user.Room.Message == nil {
		user.Room.Message = make([]json.RawMessage, 0, 1)
	}

	user.Room.Message = append(user.Room.Message, msg.Origin)

	print(user.Room.Message)

	i.BroadcastToRoom(user.Room, "chat", msg.Origin)
}

func (i *Im) joinRoom(msg *distribution.Message) {
	user, _ := msg.User().(*User)

	room := Room{}

	if err := json.Unmarshal(msg.Origin, &room); err != nil {
		log.Println(err)
		return
	}

	if room, ok := i.rooms[room.ID]; ok {
		answer, err := i.distribution.Tracker.Track(room.Master.Conn, &distribution.Message{
			Event:   "askPlayInfo",
			Payload: map[int]int{},
		})

		if err != nil {
			log.Println(err)
			return
		}

		room.AddUser(user)
		_ = i.distribution.Reply(true, msg, answer.Origin)
		i.BroadcastToRoom(room, "joinRoom", user)
		return
	}
	_ = i.distribution.Reply(false, msg, &Response{Message: "房间不存在"})
}

func (i *Im) roomInfo(msg *distribution.Message) {
	var room Room
	if err := json.Unmarshal(msg.Origin, &room); err != nil {
		log.Println(err)
		return
	}
	if room, ok := i.rooms[room.ID]; ok {
		if err := i.distribution.Reply(true, msg, room); err != nil {
			log.Println(err)
		}
		return
	}
	_ = i.distribution.Reply(false, msg, Response{Message: "房间不存在"})
}
