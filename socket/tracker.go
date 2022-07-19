package socket

type Tracker struct {
	events map[string]chan *Message
}

func newTracker() *Tracker {
	return &Tracker{
		events: make(map[string]chan *Message),
	}
}

func (t *Tracker) Track(id string) chan *Message {
	t.events[id] = make(chan *Message)
	return t.events[id]
}

func (t *Tracker) Listen(message *Message) {
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
