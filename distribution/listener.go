package distribution

type Middleware func(message *Message) bool

type Action func(message *Message)

type Listener struct {
	Middlewares []Middleware
	Action      Action
}
