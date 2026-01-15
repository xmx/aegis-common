package muxconn

import (
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

type goQUICConn struct {
	stm *quic.Stream
	mst *goQUIC
}

func (c *goQUICConn) Read(b []byte) (int, error)         { return c.stm.Read(b) }
func (c *goQUICConn) Write(b []byte) (int, error)        { return c.stm.Write(b) }
func (c *goQUICConn) Close() error                       { return c.stm.Close() }
func (c *goQUICConn) LocalAddr() net.Addr                { return c.mst.Addr() }
func (c *goQUICConn) RemoteAddr() net.Addr               { return c.mst.RemoteAddr() }
func (c *goQUICConn) SetDeadline(t time.Time) error      { return c.stm.SetDeadline(t) }
func (c *goQUICConn) SetReadDeadline(t time.Time) error  { return c.stm.SetReadDeadline(t) }
func (c *goQUICConn) SetWriteDeadline(t time.Time) error { return c.stm.SetWriteDeadline(t) }
