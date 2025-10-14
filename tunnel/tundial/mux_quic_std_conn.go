package tundial

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/quic"
)

type quicStdConn struct {
	stm     *quic.Stream
	laddr   net.Addr
	raddr   net.Addr
	traffic *trafficCounter
	parent  context.Context
}

func (q *quicStdConn) Read(b []byte) (int, error) {
	n, err := q.stm.Read(b)
	q.traffic.incrRX(n)
	return n, err
}

func (q *quicStdConn) Write(b []byte) (int, error) {
	n, err := q.stm.Write(b)
	q.traffic.incrTX(n)
	return n, err
}

func (q *quicStdConn) Close() error {
	return q.stm.Close()
}

func (q *quicStdConn) LocalAddr() net.Addr {
	return q.laddr
}

func (q *quicStdConn) RemoteAddr() net.Addr {
	return q.raddr
}

func (q *quicStdConn) SetDeadline(t time.Time) error {
	ctx := q.withDeadline(t)
	_ = q.setReadDeadline(ctx)
	return q.setWriteDeadline(ctx)
}

func (q *quicStdConn) SetReadDeadline(t time.Time) error {
	ctx := q.withDeadline(t)
	return q.setReadDeadline(ctx)
}

func (q *quicStdConn) SetWriteDeadline(t time.Time) error {
	ctx := q.withDeadline(t)
	return q.setWriteDeadline(ctx)
}

func (q *quicStdConn) setReadDeadline(ctx context.Context) error {
	q.stm.SetReadContext(ctx)
	return nil
}

func (q *quicStdConn) setWriteDeadline(ctx context.Context) error {
	q.stm.SetWriteContext(ctx)
	return nil
}

func (q *quicStdConn) withDeadline(t time.Time) context.Context {
	//goland:noinspection GoVetLostCancel
	ctx, _ := context.WithDeadline(q.parent, t)
	return ctx
}
