package distribution

import (
	"errors"
	"github.com/google/uuid"
	"hourglass-socket/socket"
	"time"
)

var TimeoutErr = errors.New("request Timout")

type Tracker struct {
	events map[string]chan *Message
}

func (t *Tracker) MakeChannel(id string) chan *Message {
	t.events[id] = make(chan *Message)
	return t.events[id]
}

// Handle 处理 reply
func (t *Tracker) Handle(message *Message) {
	if ch, ok := t.events[message.ID]; ok {
		ch <- message
		delete(t.events, message.ID)
		close(ch)
	}
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

func (t *Tracker) Track(conn *socket.Connect, message *Message) (*Message, error) {
	message.ID = uuid.New().String()
	
	data, err := message.JsonEncode()
	if err != nil {
		return nil, err
	}

	if err := conn.Emit(data); err != nil {
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
