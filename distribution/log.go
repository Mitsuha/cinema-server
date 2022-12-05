package distribution

import (
	"encoding/json"
	"log"
	"strings"
)

type ShowConnectLog struct{}

func (l *ShowConnectLog) Boot(distribution *Distribution) {
	distribution.Hook("connect", l.connect)
	distribution.Hook("disconnect", l.disconnect)
	distribution.Hook("disconnect", l.disconnect)
}

func (l *ShowConnectLog) connect(_ *Message) bool {
	log.Println("connected")

	return true
}

func (l *ShowConnectLog) disconnect(m *Message) bool {
	log.Printf("disconnect %s \n", m.Payload.(map[string]string)["message"])
	return true
}

func (l *ShowConnectLog) message(m *Message) bool {
	raw, _ := json.Marshal(m)

	if strings.Contains(string(raw), "createRoom") {
		log.Printf("recive: \t %s", "createRoom")
	} else {
		log.Printf("recive: \t %s", raw)
	}

	return true
}
