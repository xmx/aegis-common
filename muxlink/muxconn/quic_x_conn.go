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
	stm    *quic.Stream
	mst    *xQUIC
	lrw    io.ReadWriter
	closed atomic.Bool
	cancel context.CancelCauseFunc
}

func (x *xQUICConn) Read(b []byte) (int, error) {
	n, err := x.lrw.Read(b)
	x.mst.traffic.incrRX(n)

	return n, err
}

func (x *xQUICConn) Write(b []byte) (int, error) {
	n, err := x.lrw.Write(b)
	_ = x.stm.Flush()
	x.mst.traffic.incrTX(n)

	return n, err
}

func (x *xQUICConn) Close() error {
	if !x.closed.CompareAndSwap(false, true) {
		return net.ErrClosed
	}

	err := x.stm.Close()
	x.cancel(net.ErrClosed)
	x.mst.streams.closeOne()

	return err
}

func (x *xQUICConn) LocalAddr() net.Addr  { return x.mst.Addr() }
func (x *xQUICConn) RemoteAddr() net.Addr { return x.mst.RemoteAddr() }

func (x *xQUICConn) SetDeadline(t time.Time) error {
	ctx := x.withContext(t)
	x.stm.SetReadContext(ctx)
	x.stm.SetWriteContext(ctx)

	return nil
}

func (x *xQUICConn) SetReadDeadline(t time.Time) error {
	x.stm.SetReadContext(x.withContext(t))

	return nil
}

func (x *xQUICConn) SetWriteDeadline(t time.Time) error {
	x.stm.SetWriteContext(x.withContext(t))

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
