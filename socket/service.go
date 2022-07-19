package socket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

var TimeoutErr = errors.New("request Timout")

type Listener func(*Message)

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
		fmt.Println("connected")
		for true {
			_, msgBytes, err := conn.Conn.ReadMessage()

			fmt.Println(string(msgBytes))

			if err != nil {
				_ = conn.Close()

				s.Trigger(&Message{Event: "disconnect", Conn: conn})
				return
			}

			var message = Message{Conn: conn}

			if err := json.Unmarshal(msgBytes, &message); err != nil {
				log.Println(err)
				continue
			}

			s.Trigger(&message)
		}

	}(conn)
}

// Trigger 触发监听者
func (s *Service) Trigger(msg *Message) {
	if listeners, ok := s.listener[msg.Event]; ok {
		for _, listener := range listeners {
			listener(msg)
		}
	}
}

func (s *Service) Emit(conn *Connect, message *Message) error {
	bytes, err := message.JsonEncode()
	fmt.Printf("send: %s\n", bytes)

	if err != nil {
		return err
	}
	return conn.Conn.WriteMessage(websocket.TextMessage, bytes)
}

func (s *Service) Request(conn *Connect, event string, payload interface{}) (error, *Message) {
	message := Message{
		ID:      uuid.New().String(),
		Event:   event,
		Payload: payload,
	}
	if err := s.Emit(conn, &message); err != nil {
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
