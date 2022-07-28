package distribution

import (
	"encoding/json"
	"hourglass-socket/socket"
	"log"
)

type Distribution struct {
	Listeners map[string][]*Listener
	Tracker   *Tracker
}

func Listen(service *socket.Service) *Distribution {
	var distributor = Distribution{
		Tracker: &Tracker{
			events: make(map[string]chan *Message),
		},
		Listeners: map[string][]*Listener{},
	}

	service.Listen("connect", distributor.OnConnect)
	service.Listen("disconnect", distributor.onDisconnect)
	service.Listen("message", distributor.onMessage)

	return &distributor
}

func (d *Distribution) OnConnect(conn *socket.Connect, _ []byte) {
	d.Trigger(&Message{
		Event:   "connect",
		Conn:    conn,
	})
}

func (d *Distribution) onDisconnect(conn *socket.Connect, _ []byte) {
	d.Trigger(&Message{
		Event:   "disconnect",
		Conn:    conn,
	})
}

func (d *Distribution) onMessage(conn *socket.Connect, data []byte) {
	var msg = &Message{Conn: conn}

	if err := json.Unmarshal(data, msg); err != nil {
		log.Println(err)
		return
	}
	if msg.Event == "reply" {
		d.Tracker.Handle(msg)
	} else {
		d.Trigger(msg)
	}
}

func (d *Distribution) Register(event string, listener *Listener) {
	if _, ok := d.Listeners[event]; ok {
		d.Listeners[event] = append(d.Listeners[event], listener)
	} else {
		d.Listeners[event] = []*Listener{listener}
	}
}

func (d *Distribution) RegisterMany(listeners map[string]*Listener) {
	for event, listener := range listeners{
		d.Register(event,listener)
	}
}

// Trigger 触发监听者
func (d *Distribution) Trigger(message *Message) {
	if listeners, ok := d.Listeners[message.Event]; ok {
		for _, listener := range listeners {
			var jumpOver bool
			if listener.Middlewares != nil{
				for _, middleware := range listener.Middlewares {
					if !middleware(message) {
						jumpOver = true
						break
					}
				}
			}
			if !jumpOver {
				listener.Action(message)
			}
		}
	}
}
