package socket

import (
	"github.com/gorilla/websocket"
)

type Connect struct {
	Attach interface{}
	Conn   *websocket.Conn
	Service *Service
	Online bool
}

func (c *Connect) Emit(data []byte) error {
	return c.Service.Emit(c, data)
}

func (c *Connect) Close() error {
	c.Online = false
	return c.Conn.Close()
}
