package model

// IncomingMessage - входящее сообщение от клиента
type IncomingMessage struct {
	From string
	Text string
}

// OutgoingMessage - исходящее сообщение для клиента (бизнес-модель)
type OutgoingMessage struct {
	Name    string
	Text    string
	Time    string
	Private bool
}
