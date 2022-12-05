package distribution

type Service interface {
	Boot(distribution *Distribution)
	Received(message *Message)
}

type Plugin interface {
	Boot(distribution *Distribution)
}
