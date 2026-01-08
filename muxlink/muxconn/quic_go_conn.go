package muxconn

import (
	"net"

	"github.com/quic-go/quic-go"
)

type goQUICConn struct {
	*quic.Stream
	mst *goQUIC
}

func (c *goQUICConn) LocalAddr() net.Addr  { return c.mst.Addr() }
func (c *goQUICConn) RemoteAddr() net.Addr { return c.mst.RemoteAddr() }
