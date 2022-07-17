package socket

import (
	"github.com/gorilla/websocket"
)

type Connect struct {
	User   interface{}
	Conn   *websocket.Conn
	Online bool
}

func (c *Connect) Emit(message *Message) error {
	return c.Conn.WriteJSON(message)
}
