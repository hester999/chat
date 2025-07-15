package utils

import "strings"

func IsExitCommand(msg string) bool {
	str := strings.SplitN(msg, " ", 2)

	if str[0] == "/exit" {
		return true
	}
	return false
}
