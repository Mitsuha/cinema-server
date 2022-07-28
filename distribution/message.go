package distribution

import (
    "encoding/json"
    "hourglass-socket/model"
    "hourglass-socket/socket"
)

type Message struct {
    ID      string          `json:"id"`
    Event   string          `json:"event"`
    Payload interface{}     `json:"-"`
    Origin  json.RawMessage `json:"payload"`
    Conn    *socket.Connect `json:"-"`
    User    *model.User     `json:"-"`
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
    
    return json.Marshal(m)
}
