# Простой Makefile для сборки и тестирования

.PHONY: all server client test

all: server client

server:
	go build -o bin/server ./server/cmd/server.go

client:
	go build -o bin/client ./client/cmd/client.go

test:
	go test ./... 