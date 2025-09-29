package tundial

import (
	"net"
	"time"

	"github.com/xtaci/smux"
)

type smuxConn struct {
	stream  *smux.Stream
	traffic *trafficCounter
}

func (sc *smuxConn) Read(b []byte) (int, error) {
	n, err := sc.stream.Read(b)
	sc.traffic.incrRX(n)
	return n, err
}

func (sc *smuxConn) Write(b []byte) (int, error) {
	n, err := sc.stream.Write(b)
	sc.traffic.incrTX(n)
	return n, err
}

func (sc *smuxConn) Close() error {
	return sc.stream.Close()
}

func (sc *smuxConn) LocalAddr() net.Addr {
	return sc.stream.LocalAddr()
}

func (sc *smuxConn) RemoteAddr() net.Addr {
	return sc.stream.RemoteAddr()
}

func (sc *smuxConn) SetDeadline(t time.Time) error {
	return sc.stream.SetDeadline(t)
}

func (sc *smuxConn) SetReadDeadline(t time.Time) error {
	return sc.stream.SetReadDeadline(t)
}

func (sc *smuxConn) SetWriteDeadline(t time.Time) error {
	return sc.stream.SetWriteDeadline(t)
}
