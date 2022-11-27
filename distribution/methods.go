package distribution

import (
	"hourglass-socket/socket"
)

func (d *Distribution) Reply(success bool, msg *Message, message interface{}) error {
	m := Message{
		ID:      msg.ID,
		Success: success,
		Event:   "reply",
		Payload: message,
	}

	if bytes, err := m.JsonEncode(); err != nil {
		return err
	} else {
		return d.Socket.EmitRaw(msg.Conn, bytes)
	}
}

func (d *Distribution) Send(conn *socket.Connect, event string, message interface{}) error {
	m := &Message{
		Event:   event,
		Payload: message,
	}

	if bytes, err := m.JsonEncode(); err != nil {
		return err
	} else {
		return d.Socket.EmitRaw(conn, bytes)
	}
}
