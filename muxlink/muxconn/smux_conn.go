package muxconn

import (
	"net"
	"time"

	"github.com/xtaci/smux"
)

type xtaciConn struct {
	stm *smux.Stream
	mst *xtaciSMUX
}

func (x *xtaciConn) Read(b []byte) (int, error) {
	n, err := x.stm.Read(b)
	x.mst.traffic.incrRX(n)

	return n, err
}

func (x *xtaciConn) Write(b []byte) (int, error) {
	n, err := x.stm.Write(b)
	x.mst.traffic.incrTX(n)

	return n, err
}

func (x *xtaciConn) Close() error                       { return x.stm.Close() }
func (x *xtaciConn) LocalAddr() net.Addr                { return x.mst.Addr() }
func (x *xtaciConn) RemoteAddr() net.Addr               { return x.mst.RemoteAddr() }
func (x *xtaciConn) SetDeadline(t time.Time) error      { return x.stm.SetDeadline(t) }
func (x *xtaciConn) SetReadDeadline(t time.Time) error  { return x.stm.SetReadDeadline(t) }
func (x *xtaciConn) SetWriteDeadline(t time.Time) error { return x.stm.SetWriteDeadline(t) }
