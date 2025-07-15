package utils

import (
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
