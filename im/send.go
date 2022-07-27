package im

import (
	"fmt"
	"hourglass-socket/distribution"
	"hourglass-socket/socket"
)

func (i *Im) Reply(success bool, msg *distribution.Message, message interface{}) error {
	return i.ws.Emit(msg.Conn, &distribution.Message{
		ID:      msg.ID,
		Success: success,
		Event:   "reply",
		Payload: message,
	})
}

func (i *Im) Send(conn *socket.Connect, event string, message interface{}) error {
	return i.ws.Emit(conn, &distribution.Message{
		Event:   event,
		Payload: message,
	})

}

func (i *Im) BroadcastToRoom(room *Room, event string, message interface{}) []error {
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
