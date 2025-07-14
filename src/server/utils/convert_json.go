package utils

import (
	"chat/server/internal/model"
	"encoding/json"
)

// JsonToStruct парсит JSON в любую структуру
func JsonToStruct(text string, obj interface{}) error {
	err := json.Unmarshal([]byte(text), obj)
	if err != nil {
		return err
	}
	return nil
}

// ClientMessageToJsonStr конвертирует бизнес-модель в JSON через DTO
func ClientMessageToJsonStr(data model.OutgoingMessage) []byte {
	// DTO для отправки
	var dto struct {
		Name    string `json:"name"`
		Text    string `json:"text"`
		Time    string `json:"time"`
		Private bool   `json:"private,omitempty"`
	}

	// Конвертируем бизнес-модель в DTO
	dto.Name = data.Name
	dto.Text = data.Text
	dto.Time = data.Time
	dto.Private = data.Private

	str, _ := json.Marshal(dto)
	return str
}
