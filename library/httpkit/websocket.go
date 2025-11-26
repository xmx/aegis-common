package httpkit

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func NewWebsocketUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout:  10 * time.Second,
		CheckOrigin:       func(*http.Request) bool { return true },
		EnableCompression: true,
	}
}
