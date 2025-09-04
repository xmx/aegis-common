package transport

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/quic"
)

type quicConn struct {
	stm   *quic.Stream
	laddr net.Addr
	raddr net.Addr
}

func (qc *quicConn) Read(b []byte) (int, error) {
	return qc.stm.Read(b)
}

func (qc *quicConn) Write(b []byte) (int, error) {
	return qc.stm.Write(b)
}

func (qc *quicConn) Close() error {
	return qc.stm.Close()
}

func (qc *quicConn) LocalAddr() net.Addr {
	return qc.laddr
}

func (qc *quicConn) RemoteAddr() net.Addr {
	return qc.raddr
}

func (qc *quicConn) SetDeadline(t time.Time) error {
	err := qc.SetReadDeadline(t)
	if exx := qc.SetWriteDeadline(t); exx != nil {
		err = exx
	}

	return err
}

func (qc *quicConn) SetReadDeadline(t time.Time) error {
	//goland:noinspection GoVetLostCancel
	ctx, _ := context.WithDeadline(context.Background(), t)
	qc.stm.SetReadContext(ctx)

	return nil
}

func (qc *quicConn) SetWriteDeadline(t time.Time) error {
	//goland:noinspection GoVetLostCancel
	ctx, _ := context.WithDeadline(context.Background(), t)
	qc.stm.SetWriteContext(ctx)

	return nil
}
