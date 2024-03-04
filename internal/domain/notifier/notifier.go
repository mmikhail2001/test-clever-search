package notifier

import "github.com/gorilla/websocket"

type Notify struct {
	Event string
	// TODO: здесь string, а в file uuid.UUID
	UserID  string
	FileURL string
}

type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan Notify
}
