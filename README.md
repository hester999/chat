# Chat — консольный многопротокольный чат на Go

## Описание

Chat —  консольный чат-сервер и клиент на Go, поддерживающий одновременную работу по протоколам HTTP/WebSocket, TCP и UDP. Проект реализует регистрацию пользователей, публичные и приватные сообщения, обработку ошибок и простую архитектуру с единым интерфейсом для всех транспортов.

---

## Реализовано

- **HTTP/WebSocket чат** — обмен сообщениями через WebSocket, поддержка приватных и публичных сообщений.
- **TCP чат** — классический чат по TCP, поддержка приватных и публичных сообщений.
- **UDP чат** — обмен сообщениями по UDP, поддержка приватных и публичных сообщений.
- **Приватные сообщения (whisper)** — отправка личных сообщений по имени пользователя.
- **Публичные сообщения (broadcast)** — рассылка всем пользователям.
- **Регистрация пользователей** — уникальные имена, проверка на дублирование.
- **Обработка ошибок** — ошибки регистрации, некорректный JSON, отсутствие получателя и др.
- **Унифицированная архитектура** — все транспорты реализуют общий интерфейс.

---

## Основные команды (для клиента)

- `register <username>` — регистрирует пользователя с первым подключением, отправляется с первым подключеним в JSON формате 
- `broadcast <message>` — отправить публичное сообщение по умолчанию
- `/whisper <username> <message>` — отправить приватное сообщение
- `/exit` — выйти из чата


---

## Архитектура и интерфейсы

В основе лежит единый интерфейс транспорта:

```go
// server/internal/app/chat_server.go

type Transport interface {
    Start(address string) error
    Stop() error
    BroadcastMessage(msg model.IncomingMessage) error
    SendPrivateMessage(msg model.IncomingMessage) error
}
```

- Каждый протокол (HTTP, TCP, UDP) реализует этот интерфейс.
- Сервер работает только с этим интерфейсом, не зная деталей реализации.
- Для передачи сообщений используется универсальный DTO (JSON-строка в поле `Text`).

---

## Сборка и запуск

### Требования
- Go 1.20+
- make (опционально, для удобства)

### Сборка

```sh
make build
```

- Собирает сервер и клиент для всех протоколов.
- Можно собирать вручную через `go build` в соответствующих папках.

### Запуск

```sh
make run-server
make run-client-http
make run-client-tcp
make run-client-udp
```

- Или вручную:
  - Сервер: `go run ./src/server/cmd/tcp/server.go` (или http/udp)
  - Клиент: `go run ./src/client/cmd/client.go` (или client_tcp.go, client_udp.go)

### Флаги

- Адрес сервера, имя пользователя и другие параметры задаются через флаги командной строки:
  -  -p - тип протокола (***tcp, udp, http***) на котором запускается сервер и клиент
  -  -port -  порт на котором запускается сервер и клиент (по умолчанию ***4545***)
  -  -ip - адрес на котором запускается сервер и клиент (по умолчанию ***127.0.0.1***)
  
  ````
  // пример запуска сервера и клиента  на localhost:5445 по протоколу tcp
  go run client -p tcp
  go run server -p tcp
  ````

---

## Тесты

- Все тесты находятся в папках `src/client/test/` и `src/server/test/`.
- Покрывают:
  - Логику DTO и сериализации
  - Проверку регистрации и ошибок
  - Юнит-тесты для утилит и моделей
  - Моки для интерфейсов

### Запуск тестов

```sh
make test
```
или
```sh
cd src/server && go test ./...
cd src/client && go test ./...
```

---

