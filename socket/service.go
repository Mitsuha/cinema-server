package socket

import (
	"github.com/gorilla/websocket"
	"log"
)

type Listener func(*Connect, []byte)

type Service struct {
	Connects  []*Connect
	Listeners map[string][]Listener
}

func New() *Service {
	service := &Service{
		Connects:  make([]*Connect, 0, 100),
		Listeners: map[string][]Listener{},
	}

	return service
}

func (s *Service) HandleConn(conn *websocket.Conn) {
	s.newHandler(&Connect{
		Conn:   conn,
		Online: true,
		Service: s,
	})
}

func (s *Service) newHandler(conn *Connect) {
	go func(conn *Connect) {
		log.Println("connected")
		for true {
			_, msg, err := conn.Conn.ReadMessage()

			log.Printf("recive: \t %s", msg)

			if err != nil {
				log.Println(err)
				_ = conn.Close()

				s.Trigger(conn, "disconnect", nil)
				return
			}

			s.Trigger(conn, "message", msg)
		}

	}(conn)
}

func (s *Service) Listen(event string, listener Listener) {
	if _, ok := s.Listeners[event]; ok {
		s.Listeners[event] = append(s.Listeners[event], listener)
	} else {
		s.Listeners[event] = []Listener{listener}
	}
}

// Trigger 触发监听者
func (s *Service) Trigger(conn *Connect, event string, data []byte) {
	if listeners, ok := s.Listeners[event]; ok {
		for _, listener := range listeners {
			listener(conn, data)
		}
	}
}

func (s *Service) Emit(conn *Connect, data []byte) error {
	log.Printf("send: \t%s\n", data)
	
	return conn.Conn.WriteMessage(websocket.TextMessage, data)
}
