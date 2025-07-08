package model

type IncomingMessage struct {
	From string
	Text string
}

type OutgoingMessage struct {
	Name string `json:"name"`
	Text string `json:"text"`
	Time string `json:"time"`
}
