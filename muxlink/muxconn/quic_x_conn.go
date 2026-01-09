package muxconn

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/quic"
)

type xQUICConn struct {
	stm *quic.Stream
	mst *xQUIC
}

func (x *xQUICConn) Read(b []byte) (int, error) {
	n, err := x.stm.Read(b)
	x.mst.traffic.incrRX(n)

	return n, err
}

func (x *xQUICConn) Write(b []byte) (int, error) {
	n, err := x.stm.Write(b)
	_ = x.stm.Flush()
	x.mst.traffic.incrTX(n)

	return n, err
}

func (x *xQUICConn) Close() error {
	return x.stm.Close()
}

func (x *xQUICConn) LocalAddr() net.Addr {
	return x.mst.Addr()
}

func (x *xQUICConn) RemoteAddr() net.Addr {
	return x.mst.RemoteAddr()
}

func (x *xQUICConn) SetDeadline(t time.Time) error {
	_ = x.SetReadDeadline(t)
	_ = x.SetWriteDeadline(t)

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
