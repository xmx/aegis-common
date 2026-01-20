package muxconn

import (
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/hashicorp/yamux"
)

type yamuxConn struct {
	stm    *yamux.Stream
	mst    *yamuxMUX
	lrw    io.ReadWriter
	closed atomic.Bool
	cancel context.CancelCauseFunc
}

func (c *yamuxConn) Read(b []byte) (int, error) {
	n, err := c.lrw.Read(b)
	c.mst.traffic.incrRX(n)

	return n, err
}

func (c *yamuxConn) Write(b []byte) (int, error) {
	n, err := c.lrw.Write(b)
	c.mst.traffic.incrTX(n)

	return n, err
}

func (c *yamuxConn) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return net.ErrClosed
	}

	err := c.stm.Close()
	c.cancel(net.ErrClosed)
	c.mst.streams.closeOne()

	return err
}

func (c *yamuxConn) LocalAddr() net.Addr                { return c.stm.LocalAddr() }
func (c *yamuxConn) RemoteAddr() net.Addr               { return c.stm.RemoteAddr() }
func (c *yamuxConn) SetDeadline(t time.Time) error      { return c.stm.SetDeadline(t) }
func (c *yamuxConn) SetReadDeadline(t time.Time) error  { return c.stm.SetReadDeadline(t) }
func (c *yamuxConn) SetWriteDeadline(t time.Time) error { return c.stm.SetWriteDeadline(t) }
