package tundial

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

	// Transferred 数据传输字节数。
	Transferred() (rx, tx uint64)
}

type trafficCounter struct {
	rx, tx atomic.Uint64
}

func (t *trafficCounter) incrRX(n int) uint64 {
	if n < 0 {
		n = 0
	}

	return t.rx.Add(uint64(n))
}

func (t *trafficCounter) incrTX(n int) uint64 {
	if n < 0 {
		n = 0
	}

	return t.tx.Add(uint64(n))
}

func (t *trafficCounter) load() (uint64, uint64) {
	rx := t.rx.Load()
	tx := t.tx.Load()

	return rx, tx
}
