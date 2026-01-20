package muxconn

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
	"golang.org/x/time/rate"
)

func NewQUICgo(parent context.Context, conn *quic.Conn) Muxer {
	if parent == nil {
		parent = context.Background()
	}

	return &goQUIC{
		conn:    conn,
		limiter: newUnlimit(),
		streams: new(streamStat),
		parent:  parent,
	}
}

type goQUIC struct {
	conn    *quic.Conn
	limiter *rateLimiter
	streams *streamStat
	parent  context.Context
}

func (q *goQUIC) Accept() (net.Conn, error)              { return q.newConn(q.conn.AcceptStream(q.parent)) }
func (q *goQUIC) Open(context.Context) (net.Conn, error) { return q.newConn(q.conn.OpenStream()) }
func (q *goQUIC) Close() error                           { return q.conn.CloseWithError(0, "") }
func (q *goQUIC) Addr() net.Addr                         { return q.conn.LocalAddr() }
func (q *goQUIC) RemoteAddr() net.Addr                   { return q.conn.RemoteAddr() }
func (q *goQUIC) Library() (string, string)              { return "quic", "github.com/quic-go/quic-go" }
func (q *goQUIC) Limit() rate.Limit                      { return q.limiter.Limit() }
func (q *goQUIC) SetLimit(bps rate.Limit)                { q.limiter.SetLimit(bps) }
func (q *goQUIC) NumStreams() (int64, int64)             { return q.streams.NumStreams() }

func (q *goQUIC) Traffic() (uint64, uint64) {
	stat := q.conn.ConnectionStats()
	return stat.BytesReceived, stat.BytesSent
}

func (q *goQUIC) newConn(stm *quic.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(q.parent)
	lrw := q.limiter.newReadWriter(ctx, stm)
	q.streams.openOne()

	conn := &goQUICConn{
		stm:    stm,
		mst:    q,
		lrw:    lrw,
		cancel: cancel,
	}

	return conn, nil
}
