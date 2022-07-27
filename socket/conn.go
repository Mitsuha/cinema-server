package socket

import (
	"github.com/gorilla/websocket"
	"hourglass-socket/distribution"
)

type Connect struct {
	Attach interface{}
	Conn   *websocket.Conn
	Online bool
}

func (c *Connect) Emit(message *distribution.Message) error {
	return c.Conn.WriteJSON(message)
}

func (c *Connect) Close() error {
	c.Online = false
	return c.Conn.Close()
}
