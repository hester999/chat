package dto

type HTTPMessageDTO struct {
	Type string `json:"type"`           // "register", "exit", "broadcast", "whisper"
	Name string `json:"name"`           // Имя отправителя
	Text string `json:"text,omitempty"` // Текст сообщения
	Time string `json:"time,omitempty"` // Время
	Dst  string `json:"dst,omitempty"`  // Для whisper: имя получателя
}
