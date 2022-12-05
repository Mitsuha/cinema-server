package distribution

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	ID      string          `json:"id"`
	Event   string          `json:"event"`
	Success bool            `json:"success"`
	Payload interface{}     `json:"-"`
	Origin  json.RawMessage `json:"payload"`
}

func (m *Message) JsonEncode() ([]byte, error) {
	if m.Payload != nil {
		if _, ok := m.Payload.([]byte); ok {
			m.Origin = m.Payload.([]byte)
		} else {
			if bytes, err := json.Marshal(m.Payload); err != nil {
				return nil, err
			} else {
				m.Origin = bytes
			}
		}
	}
	fmt.Println(string(m.Origin))
	return json.Marshal(m)
}
