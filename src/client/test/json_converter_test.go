package test

import (
	"chat/client/internal/model"
	"chat/client/internal/utils"
	"reflect"
	"testing"
)

func TestClientMessageToJsonStrAndBack(t *testing.T) {
	msg := model.OutgoingMessage{
		Name:    "alice",
		Text:    "hello",
		Time:    "2024-05-01 12:00:00",
		Private: true,
	}

	data := utils.ClientMessageToJsonStr(msg)

	var got model.OutgoingMessage
	err := utils.JsonToStruct(string(data), &got)
	if err != nil {
		t.Fatalf("JsonToStruct error: %v", err)
	}

	if !reflect.DeepEqual(msg, got) {
		t.Errorf("round-trip mismatch:\nwant: %+v\ngot: %+v", msg, got)
	}
}

func TestJsonToStruct_InvalidJSON(t *testing.T) {
	var got model.OutgoingMessage
	err := utils.JsonToStruct("not a json", &got)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
