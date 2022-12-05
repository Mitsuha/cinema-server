package distribution

import (
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"time"
)

var TimeoutErr = errors.New("request Timout")

type Tracker struct {
	events map[string]chan *Message
}

var tracker = Tracker{
	events: map[string]chan *Message{},
}

func DTracker() *Tracker {
	return &tracker
}

func (t *Tracker) Boot(distribution *Distribution) {
	distribution.Hook("reply", t.Handle)
}

func (t *Tracker) MakeChannel(id string) chan *Message {
	t.events[id] = make(chan *Message)
	return t.events[id]
}

// Handle 处理 reply
func (t *Tracker) Handle(message *Message) bool {
	if ch, ok := t.events[message.ID]; ok {
		ch <- message
		delete(t.events, message.ID)
		close(ch)
	} else {
		return true
	}
	return false
}

func (t *Tracker) Close(id string) {
	if ch, ok := t.events[id]; ok {
		delete(t.events, id)
		// 先删除再关闭，防止资源竞争时写一个已经关闭的 channel
		defer close(ch)
		// 非阻塞的读一下，确保 channel 内的消息清空
		select {
		case <-ch:
			return
		default:
			return
		}
	}
}

func (t *Tracker) Track(conn *websocket.Conn, message *Message) (*Message, error) {
	message.ID = uuid.New().String()

	data, err := message.JsonEncode()
	if err != nil {
		return nil, err
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return nil, err
	}

	ch := t.MakeChannel(message.ID)
	defer t.Close(message.ID)

	select {
	case response := <-ch:
		return response, nil
	case <-time.After(3 * time.Second):
		return nil, TimeoutErr
	}
}
