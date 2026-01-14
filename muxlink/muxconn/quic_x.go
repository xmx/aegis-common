package muxconn

import (
	"context"
	"errors"
	"net"
	"time"

	"golang.org/x/net/quic"
)

// NewQUICx 标准库的 quic。
//
// endpoint 在 server 端可以为空。
func NewQUICx(parent context.Context, endpoint *quic.Endpoint, conn *quic.Conn) Muxer {
	if parent == nil {
		parent = context.Background()
	}

	return &xQUIC{
		parent:   parent,
		endpoint: endpoint,
		conn:     conn,
		traffic:  new(trafficStat),
	}
}

type xQUIC struct {
	parent   context.Context
	endpoint *quic.Endpoint
	conn     *quic.Conn
	traffic  *trafficStat
}

func (x *xQUIC) Accept() (net.Conn, error) {
	stm, err := x.conn.AcceptStream(x.parent)
	if err != nil {
		return nil, err
	}

	return x.newConn(stm), nil
}

func (x *xQUIC) Close() error {
	err := x.conn.Close()
	end := x.endpoint
	if end == nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	err1 := end.Close(ctx)
	cancel()

	return errors.Join(err, err1)
}

func (x *xQUIC) Open(ctx context.Context) (net.Conn, error) {
	stm, err := x.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}

	return x.newConn(stm), nil
}

func (x *xQUIC) Addr() net.Addr             { return net.UDPAddrFromAddrPort(x.conn.LocalAddr()) }
func (x *xQUIC) RemoteAddr() net.Addr       { return net.UDPAddrFromAddrPort(x.conn.RemoteAddr()) }
func (x *xQUIC) Protocol() (string, string) { return "quic", "golang.org/x/net/quic" }
func (x *xQUIC) Traffic() (uint64, uint64)  { return x.traffic.Load() }

func (x *xQUIC) newConn(stm *quic.Stream) *xQUICConn {
	return &xQUICConn{stm: stm, mst: x}
}
