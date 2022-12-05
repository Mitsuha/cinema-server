package distribution

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"strings"
)

type ServiceCreator func(conn *websocket.Conn) Service
type Hooker func(message *Message) bool

type Distribution struct {
	serviceCreator ServiceCreator
	hooks          map[string][]Hooker
}

func New(maker ServiceCreator) *Distribution {
	var distributor = Distribution{
		serviceCreator: maker,
		hooks:          make(map[string][]Hooker),
	}
	return &distributor
}

func (d *Distribution) Enable(plugin Plugin) {
	plugin.Boot(d)
}

func (d *Distribution) Hook(event string, hooker Hooker) {
	if hookers, ok := d.hooks[event]; ok {
		d.hooks[event] = append(hookers, hooker)
	} else {
		d.hooks[event] = []Hooker{hooker}
	}
}

func (d *Distribution) TakeOver(conn *websocket.Conn) {
	var service = d.serviceCreator(conn)
	service.Boot(d)

	d.Trigger(service, &Message{Event: "connect"})
	for true {
		_, raw, err := conn.ReadMessage()

		if err != nil {
			_ = conn.Close()

			d.Trigger(service, &Message{Event: "disconnect", Payload: map[string]string{
				"message": err.Error(),
			}})
			return
		}

		if strings.Contains(string(raw), "createRoom") {
			log.Printf("recive: \t %s", "createRoom")
		} else {
			log.Printf("recive: \t %s", raw)
		}

		var message Message
		if err := json.Unmarshal(raw, &message); err == nil {
			d.Trigger(service, &message)
		}
	}
}

func (d *Distribution) Trigger(service Service, message *Message) {
	if hookers, ok := d.hooks[message.Event]; ok {
		for _, hooker := range hookers {
			if !hooker(message) {
				return
			}
		}
	}
	service.Received(message)
}
