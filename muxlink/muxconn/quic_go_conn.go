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
	master *goQUIC
	stream *quic.Stream
	limit  io.ReadWriter
	closed atomic.Bool
	cancel context.CancelCauseFunc
}

func (c *goQUICConn) Read(b []byte) (int, error)         { return c.limit.Read(b) }
func (c *goQUICConn) Write(b []byte) (int, error)        { return c.limit.Write(b) }
func (c *goQUICConn) LocalAddr() net.Addr                { return c.master.Addr() }
func (c *goQUICConn) RemoteAddr() net.Addr               { return c.master.RemoteAddr() }
func (c *goQUICConn) SetDeadline(t time.Time) error      { return c.stream.SetDeadline(t) }
func (c *goQUICConn) SetReadDeadline(t time.Time) error  { return c.stream.SetReadDeadline(t) }
func (c *goQUICConn) SetWriteDeadline(t time.Time) error { return c.stream.SetWriteDeadline(t) }

func (c *goQUICConn) Close() error {
	if !c.closed.CompareAndSwap(false, true) {
		return net.ErrClosed
	}

	c.master.streams.decr()
	err := c.stream.Close()
	c.cancel(net.ErrClosed)

	return err
}
