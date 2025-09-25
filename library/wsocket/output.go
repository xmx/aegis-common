package wsocket

import (
	"io"

	"github.com/gorilla/websocket"
)

func NewChannelWriter(ws *websocket.Conn, channel string) io.Writer {
	return &channelWriter{
		conn:    ws,
		channel: channel,
	}
}

type channelWriter struct {
	channel string
	conn    *websocket.Conn
}

func (cw *channelWriter) Write(p []byte) (int, error) {
	n := len(p)
	msg := &channelMessage{Channel: cw.channel, Message: string(p)}
	if err := cw.conn.WriteJSON(msg); err != nil {
		return 0, err
	}

	return n, nil
}

type channelMessage struct {
	Channel string `json:"channel"`
	Message string `json:"message"`
}
