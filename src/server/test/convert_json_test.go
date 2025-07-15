package test

import (
	"chat/server/utils"
	"testing"
)

type testStruct struct {
	Field string `json:"field"`
}

func TestJsonToStruct_Valid(t *testing.T) {
	input := `{"field":"value"}`
	var obj testStruct
	err := utils.JsonToStruct(input, &obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obj.Field != "value" {
		t.Errorf("expected 'value', got '%s'", obj.Field)
	}
}

func TestJsonToStruct_Invalid(t *testing.T) {
	input := `not a json`
	var obj testStruct
	err := utils.JsonToStruct(input, &obj)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
