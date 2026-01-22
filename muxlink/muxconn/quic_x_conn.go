package muxconn

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	"golang.org/x/net/quic"
)

type xQUICConn struct {
	master *xQUIC
	stream *quic.Stream
	limit  io.ReadWriter
	closed atomic.Bool
	cancel context.CancelCauseFunc
}

func (x *xQUICConn) Read(b []byte) (int, error) {
	n, err := x.limit.Read(b)
	x.master.traffic.incrRX(n)

	return n, err
}

func (x *xQUICConn) Write(b []byte) (int, error) {
	n, err := x.limit.Write(b)
	_ = x.stream.Flush()
	x.master.traffic.incrTX(n)

	return n, err
}

func (x *xQUICConn) Close() error {
	if !x.closed.CompareAndSwap(false, true) {
		return net.ErrClosed
	}

	x.master.streams.decr()
	err := x.stream.Close()
	x.cancel(net.ErrClosed)

	return err
}

func (x *xQUICConn) LocalAddr() net.Addr  { return x.master.Addr() }
func (x *xQUICConn) RemoteAddr() net.Addr { return x.master.RemoteAddr() }

func (x *xQUICConn) SetDeadline(t time.Time) error {
	ctx := x.withContext(t)
	x.stream.SetReadContext(ctx)
	x.stream.SetWriteContext(ctx)

	return nil
}

func (x *xQUICConn) SetReadDeadline(t time.Time) error {
	x.stream.SetReadContext(x.withContext(t))

	return nil
}

func (x *xQUICConn) SetWriteDeadline(t time.Time) error {
	x.stream.SetWriteContext(x.withContext(t))

	return nil
}

func (*xQUICConn) withContext(t time.Time) context.Context {
	if t.IsZero() {
		return context.Background()
	}
	//goland:noinspection GoVetLostCancel
	ctx, _ := context.WithDeadline(context.Background(), t)

	return ctx
}
