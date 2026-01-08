package muxconn

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
)

func NewQUICgo(parent context.Context, conn *quic.Conn) Muxer {
	if parent == nil {
		parent = context.Background()
	}

	return &goQUIC{
		conn:   conn,
		parent: parent,
	}
}

type goQUIC struct {
	conn   *quic.Conn
	parent context.Context
}

func (q *goQUIC) Accept() (net.Conn, error) {
	stm, err := q.conn.AcceptStream(q.parent)
	if err != nil {
		return nil, err
	}

	return q.newConn(stm), nil
}

func (q *goQUIC) Open(context.Context) (net.Conn, error) {
	stm, err := q.conn.OpenStream()
	if err != nil {
		return nil, err
	}

	return q.newConn(stm), nil
}

func (q *goQUIC) Close() error               { return q.conn.CloseWithError(0, "") }
func (q *goQUIC) Addr() net.Addr             { return q.conn.LocalAddr() }
func (q *goQUIC) RemoteAddr() net.Addr       { return q.conn.RemoteAddr() }
func (q *goQUIC) Protocol() (string, string) { return "quic", "github.com/quic-go/quic-go" }

func (q *goQUIC) Traffic() (uint64, uint64) {
	stat := q.conn.ConnectionStats()
	return stat.BytesReceived, stat.BytesSent
}

func (q *goQUIC) newConn(stm *quic.Stream) *goQUICConn {
	return &goQUICConn{Stream: stm, mst: q}
}
