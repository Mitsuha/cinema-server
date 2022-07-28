package im

import (
	"fmt"
	"hourglass-socket/distribution"
	"hourglass-socket/model"
	"hourglass-socket/socket"
)

func (i *Im) Reply(reply *distribution.Message, message interface{}) error {
	msg := distribution.Message{
		ID:      reply.ID,
		Event:   "reply",
		Payload: message,
	}
	
	if data, err := msg.JsonEncode(); err == nil {
		return reply.Conn.Emit(data)
	}else{
		return err
	}
}

func (i *Im) Send(conn *socket.Connect, event string, message interface{}) error {
	msg := distribution.Message{
		Event:   event,
		Payload: message,
	}
	
	if data, err := msg.JsonEncode(); err == nil {
		return conn.Emit(data)
	}else{
		return err
	}
}

func (i *Im) BroadcastToRoom(room *model.Room, event string, message interface{}) []error {
	errs := make([]error, 0)
	for _, user := range room.Users {
		fmt.Println(user)
		err := i.Send(user.Conn, event, message)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
