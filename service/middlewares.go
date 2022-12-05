package service

import (
	"hourglass-socket/distribution"
	"log"
)

type Handler struct {
	Middlewares []Middleware
	Action      func(message *distribution.Message)
}

type Middleware func(message *distribution.Message) bool

func (s *WatchService) Auth(message *distribution.Message) bool {
	if s.user == nil {
		err := distribution.Reply(s.conn, false, message, &Response{Message: "未登录，试试重启？"})
		if err != nil {
			log.Println(err)
		}
		return false
	}

	return true
}

func (s *WatchService) HasRoom(message *distribution.Message) bool {
	if s.user != nil && s.room != nil {
		return true
	}

	err := distribution.Emit(s.conn, "dismissCurrent", Room{})
	if err != nil {
		log.Println(err)
	}

	return false
}
