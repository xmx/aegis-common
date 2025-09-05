package wsocket

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func NewUpgrade() *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin:      func(*http.Request) bool { return true },
	}
}
