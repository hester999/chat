package utils

import (
	"chat/client/internal/model"
	"encoding/json"
)

// ClientMessageToJsonStr конвертирует бизнес-модель в JSON через DTO
func ClientMessageToJsonStr(data model.OutgoingMessage) []byte {

	var dto struct {
		Name    string `json:"name"`
		Text    string `json:"text"`
		Time    string `json:"time"`
		Private bool   `json:"private,omitempty"`
	}

	dto.Name = data.Name
	dto.Text = data.Text
	dto.Time = data.Time
	dto.Private = data.Private

	str, _ := json.Marshal(dto)
	return str
}

// JsonToStruct парсит JSON в бизнес-модель через DTO
func JsonToStruct(text string, data *model.OutgoingMessage) error {

	var dto struct {
		Name    string `json:"name"`
		Text    string `json:"text"`
		Time    string `json:"time"`
		Private bool   `json:"private,omitempty"`
	}

	err := json.Unmarshal([]byte(text), &dto)
	if err != nil {
		return err
	}

	data.Name = dto.Name
	data.Text = dto.Text
	data.Time = dto.Time
	data.Private = dto.Private

	return nil
}
