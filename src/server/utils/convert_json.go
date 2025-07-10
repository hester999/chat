package utils

import (
	"chat/server/internal/model"
	"encoding/json"
)

func JsonToStruct(text string, obj interface{}) error {
	err := json.Unmarshal([]byte(text), obj)
	if err != nil {
		return err
	}
	return nil
}

func ClientMessageToJsonStr(data model.OutgoingMessage) []byte {
	str, _ := json.Marshal(data)
	return str
}
