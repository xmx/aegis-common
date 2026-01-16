package muxconn

import (
	"context"
	"net"
	"sync/atomic"
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
}

type trafficStat struct {
	rx, tx atomic.Uint64
}

func (ts *trafficStat) Load() (rx, tx uint64) {
	return ts.rx.Load(), ts.tx.Load()
}

func (ts *trafficStat) incrRX(n int) {
	if n > 0 {
		ts.rx.Add(uint64(n))
	}
}

func (ts *trafficStat) incrTX(n int) {
	if n > 0 {
		ts.tx.Add(uint64(n))
	}
}
