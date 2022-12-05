package distribution

import (
	"github.com/gorilla/websocket"
	"log"
)

func Reply(conn *websocket.Conn, success bool, msg *Message, message interface{}) error {
	m := Message{
		ID:      msg.ID,
		Success: success,
		Event:   "reply",
		Payload: message,
	}

	if bytes, err := m.JsonEncode(); err != nil {
		return err
	} else {
		log.Println(string(bytes))

		return conn.WriteMessage(websocket.TextMessage, bytes)
	}
}

func Emit(conn *websocket.Conn, event string, message interface{}) error {
	m := &Message{
		Event:   event,
		Payload: message,
	}
	if bytes, err := m.JsonEncode(); err == nil {
		log.Println(string(bytes))

		return conn.WriteMessage(websocket.TextMessage, bytes)
	} else {
		return err
	}
}
