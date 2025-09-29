package tundial

import (
	"context"
	"net"
	"net/netip"
	"time"

	"golang.org/x/net/quic"
)

// NewStdQUIC 标准 QUIC 库还在实验阶段，不稳定。
func NewStdQUIC(parent context.Context, endpoint *quic.Endpoint, conn *quic.Conn) Muxer {
	if parent == nil {
		parent = context.Background()
	}
	toUDPAddr := func(ap netip.AddrPort) *net.UDPAddr {
		addr, port := ap.Addr(), ap.Port()

		return &net.UDPAddr{
			IP:   addr.AsSlice(),
			Port: int(port),
			Zone: addr.Zone(),
		}
	}

	return &quicStd{
		conn:     conn,
		laddr:    toUDPAddr(conn.LocalAddr()),
		raddr:    toUDPAddr(conn.RemoteAddr()),
		endpoint: endpoint,
		parent:   parent,
		traffic:  new(trafficCounter),
	}
}

type quicStd struct {
	conn     *quic.Conn
	laddr    net.Addr
	raddr    net.Addr
	endpoint *quic.Endpoint
	parent   context.Context
	traffic  *trafficCounter
}

func (q *quicStd) Accept() (net.Conn, error) {
	stm, err := q.conn.AcceptStream(q.parent)
	if err != nil {
		return nil, err
	}
	conn := q.makeConn(stm)

	return conn, nil
}

func (q *quicStd) Close() error {
	_ = q.conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return q.endpoint.Close(ctx)
}

func (q *quicStd) Addr() net.Addr {
	return q.laddr
}

func (q *quicStd) Open(ctx context.Context) (net.Conn, error) {
	stm, err := q.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	conn := q.makeConn(stm)

	return conn, nil
}

func (q *quicStd) RemoteAddr() net.Addr {
	return q.raddr
}

func (q *quicStd) Protocol() (string, string) {
	return "udp", "golang.org/x/net/quic"
}

func (q *quicStd) Transferred() (uint64, uint64) {
	return q.traffic.load()
}

func (q *quicStd) makeConn(stm *quic.Stream) *quicStdConn {
	return &quicStdConn{
		stm:     stm,
		laddr:   q.laddr,
		raddr:   q.raddr,
		traffic: q.traffic,
		parent:  q.parent,
	}
}
