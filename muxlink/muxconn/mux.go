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

	// Protocol 返回通信协议类型。
	//	- protocol: 标准的底层通信协议，如：tcp udp
	//	- subprotocol: 子协议或具体的通信协议实现，一般用于开发者识别追溯，如：github.com/quic-go/quic-go
	Protocol() (protocol, subprotocol string)

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
