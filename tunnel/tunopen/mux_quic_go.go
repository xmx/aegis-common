package tunopen

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
)

func NewQUICGo(parent context.Context, conn *quic.Conn) Muxer {
	if parent == nil {
		parent = context.Background()
	}

	return &quicGo{
		conn:    conn,
		laddr:   conn.LocalAddr(),
		raddr:   conn.RemoteAddr(),
		traffic: new(trafficCounter),
		parent:  parent,
	}
}

type quicGo struct {
	conn    *quic.Conn
	laddr   net.Addr
	raddr   net.Addr
	traffic *trafficCounter
	parent  context.Context
}

func (q *quicGo) Accept() (net.Conn, error) {
	stm, err := q.conn.AcceptStream(q.parent)
	if err != nil {
		return nil, err
	}
	conn := q.makeConn(stm)

	return conn, nil
}

func (q *quicGo) Close() error {
	return q.conn.CloseWithError(0, "")
}

func (q *quicGo) Addr() net.Addr {
	return q.laddr
}

func (q *quicGo) Open(ctx context.Context) (net.Conn, error) {
	stm, err := q.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	conn := q.makeConn(stm)

	return conn, nil
}

func (q *quicGo) RemoteAddr() net.Addr {
	return q.raddr
}

func (q *quicGo) Protocol() (string, string) {
	return "udp", "github.com/quic-go/quic-go"
}

func (q *quicGo) Traffic() (uint64, uint64) {
	return q.traffic.load()
}

func (q *quicGo) makeConn(stm *quic.Stream) *quicGoConn {
	return &quicGoConn{
		stm:     stm,
		laddr:   q.laddr,
		raddr:   q.raddr,
		traffic: q.traffic,
	}
}
