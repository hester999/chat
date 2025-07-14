package model

// OutgoingMessage - исходящее сообщение для клиента (бизнес-модель)
type OutgoingMessage struct {
	Name    string
	Text    string
	Time    string
	Private bool
}
