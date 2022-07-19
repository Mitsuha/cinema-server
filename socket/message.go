package socket

import "encoding/json"

type Message struct {
	ID      string          `json:"id"`
	Event   string          `json:"event"`
	Success bool            `json:"success"`
	Payload interface{}     `json:"-"`
	Origin  json.RawMessage `json:"payload"`
	Conn    *Connect        `json:"-"`
}

func (m *Message) JsonEncode() ([]byte, error) {
	if m.Payload != nil {
		bytes, err := json.Marshal(m.Payload)
		if err != nil {
			return nil, err
		}
		m.Origin = bytes
	}

	return json.Marshal(m)
}

func (m *Message) User() interface{} {
	return m.Conn.User
}
