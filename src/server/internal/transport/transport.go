package transport

import "chat/server/internal/model"

type Transport interface {
	Start() error
	Stop() error
	BroadcastMessage(msg model.IncomingMessage) error
	SendPrivateMessage(msg model.IncomingMessage) error
}
