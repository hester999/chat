package dto

type OutgoingMessageDto struct {
	Name    string `json:"name"`
	Text    string `json:"text"`
	Time    string `json:"time"`
	Private bool   `json:"private,omitempty"`
}
