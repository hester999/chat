package test

import (
	"chat/client/internal/model"
	"encoding/json"
	"reflect"
	"testing"
)

func TestOutgoingMessageMarshalUnmarshal(t *testing.T) {
	msg := model.OutgoingMessage{
		Name:    "alice",
		Text:    "hello",
		Time:    "2024-05-01 12:00:00",
		Private: true,
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
