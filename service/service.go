package service

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"hourglass-socket/distribution"
	"hourglass-socket/model"
	"log"
	"sync"
)

type Response struct {
	Message string `json:"message"`
}

type WatchService struct {
	distributor  *distribution.Distribution
	conn         *websocket.Conn
	user         *User
	room         *Room
	RoomService  *RoomService
	roomCreating sync.Mutex
}

func New(conn *websocket.Conn) distribution.Service {
	return &WatchService{
		conn:        conn,
		RoomService: roomService,
	}
}

func (s *WatchService) Received(message *distribution.Message) {
	var events = map[string]Handler{
		"disconnect":        {Action: s.disconnect},
		"register":          {Action: s.registerUser},
		"createRoom":        {Middlewares: []Middleware{s.Auth}, Action: s.createRoom},
		"joinRoom":          {Middlewares: []Middleware{s.Auth}, Action: s.joinRoom},
		"roomInfo":          {Middlewares: []Middleware{s.Auth}, Action: s.roomInfo},
		"leaveRoom":         {Middlewares: []Middleware{s.Auth, s.HasRoom}, Action: s.leaveRoom},
		"syncPlayList":      {Middlewares: []Middleware{s.Auth, s.HasRoom}, Action: s.syncPlayList},
		"syncEpisode":       {Middlewares: []Middleware{s.Auth, s.HasRoom}, Action: s.syncEpisode},
		"syncDuration":      {Middlewares: []Middleware{s.Auth, s.HasRoom}, Action: s.syncDuration},
		"syncSpeed":         {Middlewares: []Middleware{s.Auth, s.HasRoom}, Action: s.syncSpeed},
		"syncPlayingStatus": {Middlewares: []Middleware{s.Auth, s.HasRoom}, Action: s.syncPlayingStatus},
		"chat":              {Middlewares: []Middleware{s.Auth, s.HasRoom}, Action: s.chat},
	}

	if handler, ok := events[message.Event]; ok {
		if handler.Middlewares != nil {
			for _, middleware := range handler.Middlewares {
				if !middleware(message) {
					return
				}
			}
		}
		events[message.Event].Action(message)
	}
}

func (s *WatchService) Boot(distribution *distribution.Distribution) {
	s.distributor = distribution
}

func (s *WatchService) registerUser(message *distribution.Message) {
	var user = User{}

	if err := json.Unmarshal(message.Origin, &user); err != nil {
		s.ResponseError(err)
		return
	}

	var savedUser User
	db := model.DB.Where("phone = ?", user.Phone).First(&savedUser)

	if db.Error == gorm.ErrRecordNotFound {
		model.DB.Create(&user)
	} else {
		model.DB.Model(savedUser).Updates(user)
	}

	savedUser.Conn, s.user = s.conn, &savedUser

	_ = distribution.Reply(s.conn, true, message, user)
}

func (s *WatchService) disconnect(msg *distribution.Message) {
	if s.user != nil && s.room != nil {
		s.leaveRoom(msg)
	}
}

func (s *WatchService) createRoom(msg *distribution.Message) {
	if s.room != nil {
		s.leaveRoom(msg)
	}

	s.room = s.RoomService.Create(s.user)

	go func() {
		var room = s.room.ToModel()
		model.DB.Create(room)
		s.room.ModelID = room.ID

		select {
		case _, isOpen := <-s.room.OnDismiss:
			if !isOpen {
				s.room = nil
			}
		}
	}()

	s.room.Playlist = msg.Origin

	if err := distribution.Reply(s.conn, true, msg, s.room); err != nil {
		log.Println(err)
	}
}

func (s *WatchService) joinRoom(msg *distribution.Message) {
	room := struct {
		ID int `json:"id"`
	}{}

	if err := json.Unmarshal(msg.Origin, &room); err != nil {
		s.ResponseError(err)
		return
	}

	if room := s.RoomService.Find(room.ID); room != nil {
		answer, err := distribution.DTracker().Track(room.Master.Conn, &distribution.Message{
			Event:   "askPlayInfo",
			Payload: map[int]int{},
		})

		if err != nil {
			s.ResponseError(err)
			return
		}

		_, _ = s.RoomService.JoinRoom(room, s.user)
		s.room = room

		_ = distribution.Reply(s.conn, true, msg, answer.Origin)
		s.BroadcastToRoom(room, "joinRoom", s.user)

		/// update db
		model.DB.Table("rooms").
			Where("id = ?", room.ModelID).
			Update("audience", gorm.Expr("JSON_ARRAY_APPEND(audience, '$', ?)", s.user.ID))

		return
	}
	_ = distribution.Reply(s.conn, false, msg, &Response{Message: "房间不存在"})
}

func (s *WatchService) leaveRoom(_ *distribution.Message) {
	// dismiss room
	if s.isOwner() {
		s.BroadcastToRoom(s.room, "dismiss", s.room)
		s.RoomService.Dismiss(s.room)

		return
	} else {
		s.BroadcastToRoom(s.room, "leaveRoom", s.user)

		s.RoomService.RemoteUser(s.room, s.user)
		s.room = nil
	}
}

func (s *WatchService) roomInfo(msg *distribution.Message) {
	var room struct {
		ID int `json:"id"`
	}

	if err := json.Unmarshal(msg.Origin, &room); err != nil {
		_ = distribution.Emit(s.conn, "error", &Response{Message: err.Error()})
		return
	}

	if room := s.RoomService.Find(room.ID); room != nil {
		_ = distribution.Reply(s.conn, true, msg, room)
		return
	}
	_ = distribution.Reply(s.conn, false, msg, &Response{Message: "房间不存在"})
}

func (s *WatchService) syncPlayList(msg *distribution.Message) {
	data := struct {
		Playlist json.RawMessage `json:"playlist"`
	}{}

	if err := json.Unmarshal(msg.Origin, &data); err == nil {
		s.room.Playlist = data.Playlist
		s.BroadcastToRoom(s.room, "syncPlayList", msg.Origin)
	}
}

func (s *WatchService) syncEpisode(msg *distribution.Message) {
	var data = struct {
		Index    int               `json:"index"`
		PlayInfo []json.RawMessage `json:"playInfo"`
	}{}

	if err := json.Unmarshal(msg.Origin, &data); err == nil && len(data.PlayInfo) != 0 {
		var cursor = 0
		for _, u := range s.room.Users {
			if u.AliID != s.room.Master.AliID {
				_ = distribution.Emit(u.Conn, "syncEpisode", map[string]interface{}{
					"index":    data.Index,
					"playInfo": data.PlayInfo[cursor%len(data.PlayInfo)],
				})
				cursor++
			}
		}

		s.room.Episode, s.room.Duration = data.Index, 0
	} else if err != nil {
		log.Println(err)
	}
}

func (s *WatchService) syncDuration(msg *distribution.Message) {
	var data = struct {
		Duration int `json:"duration"`
		Time     int `json:"time"`
	}{}

	if err := json.Unmarshal(msg.Origin, &data); err != nil {
		return
	}

	// todo 服务器端计时估算
	s.room.Duration = data.Duration

	s.BroadcastWithoutMaster(s.room, "syncDuration", data)
}

func (s *WatchService) syncSpeed(msg *distribution.Message) {
	data := struct {
		Speed float32 `json:"speed"`
	}{}

	if json.Unmarshal(msg.Origin, &data) == nil {
		s.room.Speed = data.Speed

		s.BroadcastToRoom(s.room, "syncSpeed", msg.Origin)
	}

}

func (s *WatchService) syncPlayingStatus(msg *distribution.Message) {
	var data = struct {
		Playing bool `json:"playing"`
	}{}

	if err := json.Unmarshal(msg.Origin, &data); err != nil {
		return
	}

	s.room.IsPlaying = data.Playing

	s.BroadcastToRoom(s.room, "syncPlayingStatus", data)
}

func (s *WatchService) chat(msg *distribution.Message) {
	if s.room.Message == nil {
		s.room.Message = make([]model.Message, 0, 1)
	}

	var message model.Message

	if err := json.Unmarshal(msg.Origin, &message); err == nil {
		message.RoomID, message.UserID = s.room.ModelID, s.user.ID
		s.room.Message = append(s.room.Message, message)

		s.BroadcastToRoom(s.room, "chat", msg.Origin)

		model.DB.Create(&message)
	}
}

func (s *WatchService) isOwner() bool {
	return s.room.Master.AliID == s.user.AliID
}

func (s *WatchService) ResponseError(err error) {
	_ = distribution.Emit(s.conn, "error", &Response{Message: err.Error()})
}
