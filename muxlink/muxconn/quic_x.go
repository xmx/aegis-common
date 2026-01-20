package muxconn

import (
	"context"
	"errors"
	"net"
	"time"

	"golang.org/x/net/quic"
	"golang.org/x/time/rate"
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
		limiter:  newUnlimit(),
		streams:  new(streamStat),
	}
}

type xQUIC struct {
	parent   context.Context
	endpoint *quic.Endpoint
	conn     *quic.Conn
	traffic  *trafficStat
	limiter  *rateLimiter // 读写限流器
	streams  *streamStat  // stream 计数器
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

func (x *xQUIC) Accept() (net.Conn, error)                  { return x.newConn(x.conn.AcceptStream(x.parent)) }
func (x *xQUIC) Open(ctx context.Context) (net.Conn, error) { return x.newConn(x.conn.NewStream(ctx)) }
func (x *xQUIC) Addr() net.Addr                             { return net.UDPAddrFromAddrPort(x.conn.LocalAddr()) }
func (x *xQUIC) RemoteAddr() net.Addr                       { return net.UDPAddrFromAddrPort(x.conn.RemoteAddr()) }
func (x *xQUIC) Library() (string, string)                  { return "quic", "golang.org/x/net/quic" }
func (x *xQUIC) Traffic() (uint64, uint64)                  { return x.traffic.Load() }
func (x *xQUIC) Limit() rate.Limit                          { return x.limiter.Limit() }
func (x *xQUIC) SetLimit(bps rate.Limit)                    { x.limiter.SetLimit(bps) }
func (x *xQUIC) NumStreams() (int64, int64)                 { return x.streams.NumStreams() }

func (x *xQUIC) newConn(stm *quic.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	parent := x.parent
	ctx, cancel := context.WithCancelCause(parent)
	lrw := x.limiter.newReadWriter(ctx, stm)
	x.streams.openOne()

	conn := &xQUICConn{
		stm:    stm,
		mst:    x,
		lrw:    lrw,
		cancel: cancel,
	}

	return conn, nil
}
