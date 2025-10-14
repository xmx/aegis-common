package tundial

import (
	"context"
	"errors"
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

	return &quicStd{
		conn:     conn,
		endpoint: endpoint,
		parent:   parent,
		traffic:  new(trafficCounter),
	}
}

type quicStd struct {
	conn     *quic.Conn
	endpoint *quic.Endpoint // server 端可以为空。
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
	err := q.conn.Close()
	end := q.endpoint
	if end == nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err1 := end.Close(ctx)
	cancel()

	return errors.Join(err, err1)
}

func (q *quicStd) Addr() net.Addr {
	addr := q.conn.LocalAddr()
	return q.toAddr(addr)
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
	addr := q.conn.RemoteAddr()
	return q.toAddr(addr)
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
		laddr:   q.Addr(),
		raddr:   q.RemoteAddr(),
		traffic: q.traffic,
		parent:  q.parent,
	}
}

func (*quicStd) toAddr(ip netip.AddrPort) *net.UDPAddr {
	addr, port := ip.Addr(), ip.Port()
	return &net.UDPAddr{
		IP:   addr.AsSlice(),
		Port: int(port),
		Zone: addr.Zone(),
	}
}
