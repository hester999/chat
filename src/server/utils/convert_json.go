package utils

import (
	"encoding/json"
	"fmt"
	"chat/server/internal/model"
)

func JsonToStruct(text string, obj interface{}) error {
	err := json.Unmarshal([]byte(text), obj)
	fmt.Println(obj)
	if err != nil {
		return err
	}
	return nil
}

func ClientMessageToJsonStr(data model.OutgoingMessage) []byte {
	str, _ := json.Marshal(data)
	return str
}
