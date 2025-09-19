package transport

import (
	"context"
	"net"
	"net/netip"
	"time"

	"golang.org/x/net/quic"
)

// NewQUIC 服务端无需填写 endpoint
func NewQUIC(parent context.Context, conn *quic.Conn, endpoint *quic.Endpoint) Muxer {
	if parent == nil {
		parent = context.Background()
	}

	laddr := conn.LocalAddr()
	raddr := conn.RemoteAddr()

	return &quicMux{
		conn:     conn,
		endpoint: endpoint,
		parent:   parent,
		laddr:    laddr,
		raddr:    raddr,
	}
}

type quicMux struct {
	conn     *quic.Conn
	endpoint *quic.Endpoint
	parent   context.Context
	laddr    netip.AddrPort
	raddr    netip.AddrPort
}

func (q *quicMux) Open(ctx context.Context) (net.Conn, error) {
	if ctx == nil {
		ctx = q.parent
	}

	stm, err := q.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	conn := q.makeConn(stm)

	return conn, nil
}

func (q *quicMux) Accept() (net.Conn, error) {
	stm, err := q.conn.AcceptStream(q.parent)
	if err != nil {
		return nil, err
	}
	conn := q.makeConn(stm)

	return conn, nil
}

func (q *quicMux) Addr() net.Addr {
	return &quicAddr{addr: q.laddr}
}

func (q *quicMux) Close() error {
	err := q.conn.Close()
	if end := q.endpoint; end != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = end.Close(ctx)
		cancel()
		q.endpoint = nil
	}

	return err
}

func (q *quicMux) Protocol() string {
	return "udp"
}

func (q *quicMux) RemoteAddr() net.Addr {
	return &quicAddr{addr: q.raddr}
}

func (q *quicMux) makeConn(stm *quic.Stream) net.Conn {
	return &quicConn{
		stm:    stm,
		laddr:  q.laddr,
		raddr:  q.raddr,
		parent: q.parent,
	}
}
