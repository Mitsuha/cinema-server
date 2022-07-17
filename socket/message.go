package socket

import "encoding/json"

type Message struct {
	ID      string          `json:"id"`
	Event   string          `json:"event"`
	Payload interface{}     `json:"-"`
	Origin  json.RawMessage `json:"payload"`
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
