package tundial

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

type quicGoConn struct {
	stm     *quic.Stream
	laddr   net.Addr
	raddr   net.Addr
	traffic *trafficCounter
}

func (q *quicGoConn) Read(b []byte) (int, error) {
	n, err := q.stm.Read(b)
	q.traffic.incrRX(n)
	return n, err
}

func (q *quicGoConn) Write(b []byte) (int, error) {
	n, err := q.stm.Write(b)
	q.traffic.incrTX(n)
	return n, err
}

func (q *quicGoConn) Close() error                       { return q.stm.Close() }
func (q *quicGoConn) LocalAddr() net.Addr                { return q.laddr }
func (q *quicGoConn) RemoteAddr() net.Addr               { return q.raddr }
func (q *quicGoConn) SetDeadline(t time.Time) error      { return q.stm.SetDeadline(t) }
func (q *quicGoConn) SetReadDeadline(t time.Time) error  { return q.stm.SetReadDeadline(t) }
func (q *quicGoConn) SetWriteDeadline(t time.Time) error { return q.stm.SetWriteDeadline(t) }
