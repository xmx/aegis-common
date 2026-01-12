package wsocket

import (
	"encoding/json"
	"io"
	"time"

	"github.com/gorilla/websocket"
)

func CloseControl(ws *websocket.Conn, err error) error {
	var msg []byte
	if err != nil {
		msg = []byte(err.Error())
	}
	if len(msg) > 125 { // maxControlFramePayloadSize = 125
		msg = msg[:125] // 截取方式未考虑多码元字符，如遇此问题再修复。
	}

	return ws.WriteControl(websocket.CloseMessage, msg, time.Now().Add(5*time.Second))
}

func NewTTYWriter(ws *websocket.Conn, tp string) io.Writer {
	return &ttyWriter{
		ws: ws,
		tp: tp,
	}
}

type ttyWriter struct {
	ws *websocket.Conn
	tp string
}

func (t *ttyWriter) Write(p []byte) (int, error) {
	n := len(p)
	if n == 0 {
		return 0, nil
	}

	msg := &ttyResponse{Type: t.tp, Data: string(p)}
	if err := t.ws.WriteJSON(msg); err != nil {
		return 0, err
	}

	return n, nil
}

type ttyResponse struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type TypeMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (tm TypeMessage) Unmarshal(v any) error {
	return json.Unmarshal(tm.Data, v)
}
