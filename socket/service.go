package socket

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"time"
)

var TimeoutErr = errors.New("request Timout")

type Listener func(*Connect, *Message)

type Service struct {
	poll     map[uint]*Connect
	listener map[string][]Listener
	tracker  *Tracker
}

func New() *Service {
	service := &Service{
		poll:     map[uint]*Connect{},
		listener: map[string][]Listener{},
		tracker:  newTracker(),
	}

	service.Listen("reply", service.tracker.Listen)

	return service
}

func (s *Service) HandleConn(conn *websocket.Conn) {
	s.newHandler(&Connect{
		Conn:   conn,
		Online: true,
	})
}

func (s *Service) Listen(event string, listener Listener) {
	if _, ok := s.listener[event]; ok {
		s.listener[event] = append(s.listener[event], listener)
	} else {
		s.listener[event] = []Listener{listener}
	}
}

func (s *Service) newHandler(conn *Connect) {
	go func(conn *Connect) {
		for true {
			_, msgBytes, err := conn.Conn.ReadMessage()
			if err != nil {
				_ = conn.Conn.Close()
				conn.Online = false
				s.Trigger(conn, &Message{Event: "disconnect"})
				return
			}

			var message = Message{}
			if err := json.Unmarshal(msgBytes, &message); err != nil {
				print(err.Error())
				continue
			}

			s.Trigger(conn, &message)
		}

	}(conn)
}

// Trigger 触发监听者
func (s *Service) Trigger(conn *Connect, msg *Message) {
	if listeners, ok := s.listener[msg.Event]; ok {
		for _, listener := range listeners {
			listener(conn, msg)
		}
	}
}

func (s *Service) emit(conn *Connect, message *Message) error {
	print(message)
	bytes, err := message.JsonEncode()
	if err != nil {
		return err
	}
	return conn.Conn.WriteMessage(websocket.TextMessage, bytes)
}

func (s *Service) Send(conn *Connect, event string, payload interface{}) error {
	return s.emit(conn, &Message{Event: event, Payload: payload})
}

func (s *Service) Request(conn *Connect, event string, payload interface{}) (error, *Message) {
	message := Message{
		ID:      uuid.New().String(),
		Event:   event,
		Payload: payload,
	}
	if err := s.emit(conn, &message); err != nil {
		return err, nil
	}

	ch := s.tracker.Track(message.ID)
	defer s.tracker.Close(message.ID)

	select {
	case response := <-ch:
		return nil, response
	case <-time.After(3 * time.Second):
		return TimeoutErr, nil
	}
}
