package test

import (
	"chat/server/internal/model"
	"encoding/json"
	"reflect"
	"testing"
)

func TestIncomingMessageMarshalUnmarshal(t *testing.T) {
	msg := model.IncomingMessage{
		From: "bob",
		Text: "hi",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got model.IncomingMessage
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(msg, got) {
		t.Errorf("marshal/unmarshal mismatch:\nwant: %+v\ngot: %+v", msg, got)
	}
}

func TestOutgoingMessageMarshalUnmarshal(t *testing.T) {
	msg := model.OutgoingMessage{
		Name:    "alice",
		Text:    "hello",
		Time:    "2024-05-01 12:00:00",
		Private: false,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got model.OutgoingMessage
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(msg, got) {
		t.Errorf("marshal/unmarshal mismatch:\nwant: %+v\ngot: %+v", msg, got)
	}
}
