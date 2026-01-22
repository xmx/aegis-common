package muxconn

import (
	"context"
	"net"

	"golang.org/x/time/rate"
)

type Muxer interface {
	net.Listener

	// Open 打开一个双向子流。
	Open(context.Context) (net.Conn, error)

	// RemoteAddr 远端节点地址。
	RemoteAddr() net.Addr

	// Library 。
	Library() (name, module string)

	// Traffic 数据传输字节数。
	Traffic() (rx, tx uint64)

	// Limit 限速。
	Limit() rate.Limit

	SetLimit(bps rate.Limit)

	NumStreams() (cumulative, active int64)
}
