package test

import (
	"chat/server/internal/dto"
	"encoding/json"
	"reflect"
	"testing"
)

func TestHTTPMessageDTOMarshalUnmarshal(t *testing.T) {
	msg := dto.HTTPMessageDTO{
		Type: "whisper",
		Name: "alice",
		Text: "hello",
		Time: "2024-05-01 12:00:00",
		Dst:  "bob",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got dto.HTTPMessageDTO
	err = json.Unmarshal(data, &got)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(msg, got) {
		t.Errorf("marshal/unmarshal mismatch:\nwant: %+v\ngot: %+v", msg, got)
	}
}
