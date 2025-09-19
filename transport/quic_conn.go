package transport

import (
	"context"
	"net"
	"net/netip"
	"time"

	"golang.org/x/net/quic"
)

type quicConn struct {
	stm    *quic.Stream
	laddr  netip.AddrPort
	raddr  netip.AddrPort
	parent context.Context
}

func (q *quicConn) Read(b []byte) (int, error) {
	return q.stm.Read(b)
}

func (q *quicConn) Write(b []byte) (int, error) {
	n, err := q.stm.Write(b)
	if n > 0 {
		_ = q.stm.Flush()
	}

	return n, err
}

func (q *quicConn) Close() error {
	return q.stm.Close()
}

func (q *quicConn) LocalAddr() net.Addr {
	return &quicAddr{addr: q.laddr}
}

func (q *quicConn) RemoteAddr() net.Addr {
	return &quicAddr{addr: q.raddr}
}

func (q *quicConn) SetDeadline(t time.Time) error {
	//goland:noinspection GoVetLostCancel
	ctx, _ := context.WithDeadline(q.parent, t)
	q.stm.SetReadContext(ctx)
	q.stm.SetWriteContext(ctx)
	return nil
}

func (q *quicConn) SetReadDeadline(t time.Time) error {
	//goland:noinspection GoVetLostCancel
	ctx, _ := context.WithDeadline(q.parent, t)
	q.stm.SetReadContext(ctx)
	return nil
}

func (q *quicConn) SetWriteDeadline(t time.Time) error {
	//goland:noinspection GoVetLostCancel
	ctx, _ := context.WithDeadline(q.parent, t)
	q.stm.SetWriteContext(ctx)
	return nil
}

type quicAddr struct{ addr netip.AddrPort }

func (q quicAddr) Network() string { return "udp" }
func (q quicAddr) String() string  { return q.addr.String() }
