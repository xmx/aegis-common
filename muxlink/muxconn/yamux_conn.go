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
	master *yamuxSession
	stream *yamux.Stream
	limit  io.ReadWriter
	closed atomic.Bool
	cancel context.CancelCauseFunc
}

func (c *yamuxConn) Read(b []byte) (int, error) {
	n, err := c.limit.Read(b)
	c.master.traffic.incrRX(n)

	return n, err
}

func (c *yamuxConn) Write(b []byte) (int, error) {
	n, err := c.limit.Write(b)
	c.master.traffic.incrTX(n)

	return n, err
}

func (c *yamuxConn) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return net.ErrClosed
	}

	c.master.streams.decr()
	err := c.stream.Close()
	c.cancel(net.ErrClosed)

	return err
}

func (c *yamuxConn) LocalAddr() net.Addr                { return c.stream.LocalAddr() }
func (c *yamuxConn) RemoteAddr() net.Addr               { return c.stream.RemoteAddr() }
func (c *yamuxConn) SetDeadline(t time.Time) error      { return c.stream.SetDeadline(t) }
func (c *yamuxConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *yamuxConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }
