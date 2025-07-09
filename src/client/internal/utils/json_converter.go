package utils

import (
	"chat/client/internal/model"
	"encoding/json"
)

func ClientMessageToJsonStr(data model.OutgoingMessage) []byte {
	str, _ := json.Marshal(data)
	return str
}

func JsonToStruct(text string, data *model.OutgoingMessage) error {
	err := json.Unmarshal([]byte(text), data)

	if err != nil {
		return err
	}
	return nil
}
