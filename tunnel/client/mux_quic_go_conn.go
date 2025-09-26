package client

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

func (q *quicGoConn) Read(b []byte) (int, error)         { return q.stm.Read(b) }
func (q *quicGoConn) Write(b []byte) (int, error)        { return q.stm.Write(b) }
func (q *quicGoConn) Close() error                       { return q.stm.Close() }
func (q *quicGoConn) LocalAddr() net.Addr                { return q.laddr }
func (q *quicGoConn) RemoteAddr() net.Addr               { return q.raddr }
func (q *quicGoConn) SetDeadline(t time.Time) error      { return q.stm.SetDeadline(t) }
func (q *quicGoConn) SetReadDeadline(t time.Time) error  { return q.stm.SetReadDeadline(t) }
func (q *quicGoConn) SetWriteDeadline(t time.Time) error { return q.stm.SetWriteDeadline(t) }
