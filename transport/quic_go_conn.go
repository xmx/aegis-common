package transport

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

type quicGoConn struct {
	stm   *quic.Stream
	laddr net.Addr
	raddr net.Addr
}

func (qc *quicGoConn) Read(b []byte) (int, error) {
	return qc.stm.Read(b)
}

func (qc *quicGoConn) Write(b []byte) (int, error) {
	return qc.stm.Write(b)
}

func (qc *quicGoConn) Close() error {
	return qc.stm.Close()
}

func (qc *quicGoConn) LocalAddr() net.Addr {
	return qc.laddr
}

func (qc *quicGoConn) RemoteAddr() net.Addr {
	return qc.raddr
}

func (qc *quicGoConn) SetDeadline(t time.Time) error {
	return qc.stm.SetDeadline(t)
}

func (qc *quicGoConn) SetReadDeadline(t time.Time) error {
	return qc.stm.SetReadDeadline(t)
}

func (qc *quicGoConn) SetWriteDeadline(t time.Time) error {
	return qc.stm.SetWriteDeadline(t)
}
