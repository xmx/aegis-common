package wsocket

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func DefaultUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout:  10 * time.Second,
		CheckOrigin:       func(*http.Request) bool { return true },
		EnableCompression: true,
	}
}

func NewDialer(dc func(ctx context.Context, network, addr string) (net.Conn, error)) *websocket.Dialer {
	return &websocket.Dialer{
		NetDialContext:    dc,
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
	}
}
