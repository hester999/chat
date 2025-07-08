package utils

import (
	"encoding/json"
	"fmt"
)

func JsonToStruct(text string, obj interface{}) error {
	err := json.Unmarshal([]byte(text), obj)
	fmt.Println(obj)
	if err != nil {
		return err
	}
	return nil
}
