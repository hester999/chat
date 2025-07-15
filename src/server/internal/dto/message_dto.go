package dto

type MessageDTO struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Text string `json:"text,omitempty"`
	Time string `json:"time,omitempty"`
	Dst  string `json:"dst,omitempty"`
}
