package muxconn

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"
)

type goQUICConn struct {
	stm    *quic.Stream
	mst    *goQUIC
	lrw    io.ReadWriter
	closed atomic.Bool
	cancel context.CancelCauseFunc
}

func (c *goQUICConn) Read(b []byte) (int, error)         { return c.lrw.Read(b) }
func (c *goQUICConn) Write(b []byte) (int, error)        { return c.lrw.Write(b) }
func (c *goQUICConn) LocalAddr() net.Addr                { return c.mst.Addr() }
func (c *goQUICConn) RemoteAddr() net.Addr               { return c.mst.RemoteAddr() }
func (c *goQUICConn) SetDeadline(t time.Time) error      { return c.stm.SetDeadline(t) }
func (c *goQUICConn) SetReadDeadline(t time.Time) error  { return c.stm.SetReadDeadline(t) }
func (c *goQUICConn) SetWriteDeadline(t time.Time) error { return c.stm.SetWriteDeadline(t) }

func (c *goQUICConn) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return net.ErrClosed
	}

	err := c.stm.Close()
	c.cancel(net.ErrClosed)
	c.mst.streams.closeOne()

	return err
}
