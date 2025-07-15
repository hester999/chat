package test

import (
	"chat/server/utils"
	"testing"
)

func TestIsExitCommand(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"/exit", true},
		{"/exit some text", true},
		{"hello", false},
		{"/exiting", false},
		{" /exit", false},
	}

	for _, c := range cases {
		got := utils.IsExitCommand(c.input)
		if got != c.expected {
			t.Errorf("IsExitCommand(%q) = %v, want %v", c.input, got, c.expected)
		}
	}
}
