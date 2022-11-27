package im

import (
	"hourglass-socket/distribution"
	"log"
)

func (i *Im) Auth(message *distribution.Message) bool {
	user, ok := message.User().(*User)

	if user == nil || !ok {
		err := i.distribution.Reply(false, message, &Response{Message: "未登录，试试重启？"})
		if err != nil {
			log.Println(err)
		}
		return false
	}

	return true
}

func (i *Im) HasRoom(message *distribution.Message) bool {
	if user, ok := message.User().(*User); ok && user.Room != nil {
		return true
	}

	err := i.distribution.Send(message.Conn, "dismissCurrent", Room{})
	if err != nil {
		log.Println(err)
	}

	return false
}
