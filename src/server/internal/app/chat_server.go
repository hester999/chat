package app

// import "chat/internal/transport"

// type ChatServer interface {
// 	Start() error
// 	Stop() error
// 	Broadcast(msg string, from string)
// }

// // Реализация методов Start, Stop, Broadcast будет позже

// type Transport interface {
// 	// Запуск прослушивания порта/сервера (обычно в отдельной горутине)
// 	Listen() error

// 	// Канал, из которого ChatServer будет читать входящие сообщения от клиентов
// 	MessageChannel() <-chan IncomingMessage

// 	// Разослать сообщение всем клиентам (broadcast)
// 	BroadcastMessage(msg IncomingMessage) error

// 	// Отправить сообщение конкретному клиенту (по адресу)
// 	SendMessage(msg string, toAddr string) error

// 	// Завершить работу транспорта, закрыть соединения
// 	Close() error
// }
