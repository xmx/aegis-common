package client

import (
	"context"
	"net"
	"net/netip"
	"time"

	"golang.org/x/net/quic"
)

func newQUIC(parent context.Context, conn *quic.Conn) *quicStd {
	toUDPAddr := func(ap netip.AddrPort) *net.UDPAddr {
		addr, port := ap.Addr(), ap.Port()

		return &net.UDPAddr{
			IP:   addr.AsSlice(),
			Port: int(port),
			Zone: addr.Zone(),
		}
	}

	if parent == nil {
		parent = context.Background()
	}
	laddr, raddr := conn.LocalAddr(), conn.RemoteAddr()

	return &quicStd{
		conn:   conn,
		laddr:  toUDPAddr(laddr),
		raddr:  toUDPAddr(raddr),
		parent: parent,
	}
}

type quicStd struct {
	conn     *quic.Conn
	laddr    net.Addr
	raddr    net.Addr
	endpoint *quic.Endpoint
	parent   context.Context
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

func (q *quicStd) makeConn(stm *quic.Stream) *quicStdConn {
	return &quicStdConn{
		stm:    stm,
		laddr:  q.laddr,
		raddr:  q.raddr,
		parent: q.parent,
	}
}
