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

type AtomicMuxer interface {
	Muxer

	// Swap 覆盖更新底层 Muxer。
	//
	// 客户端 Open 成功后，返回的 Muxer 交给业务层代码持有，但是任何连接都存在掉线重连的问题，
	// 重连成功后，要让业务层代码无感的替换底层的连接通道。
	Swap(mux Muxer) (old Muxer)
}

func MakeAtomic(m Muxer) AtomicMuxer {
	if am, yes := m.(AtomicMuxer); yes {
		return am
	}
	am := new(safeMuxer)
	_ = am.Swap(m)

	return am
}

type safeMuxer struct {
	ptr atomic.Pointer[Muxer]
}

func (sm *safeMuxer) Accept() (net.Conn, error)                  { return sm.load().Accept() }
func (sm *safeMuxer) Close() error                               { return sm.load().Close() }
func (sm *safeMuxer) Addr() net.Addr                             { return sm.load().Addr() }
func (sm *safeMuxer) Open(ctx context.Context) (net.Conn, error) { return sm.load().Open(ctx) }
func (sm *safeMuxer) RemoteAddr() net.Addr                       { return sm.load().RemoteAddr() }
func (sm *safeMuxer) Protocol() (string, string)                 { return sm.load().Protocol() }
func (sm *safeMuxer) Transferred() (rx, tx uint64)               { return sm.load().Transferred() }

func (sm *safeMuxer) Swap(mux Muxer) Muxer {
	if mux == nil {
		panic("nil muxer is not allowed")
	}

	if old := sm.ptr.Swap(&mux); old != nil {
		return *old
	}

	return nil
}

func (sm *safeMuxer) load() Muxer {
	if m := sm.ptr.Load(); m != nil {
		return *m
	}

	panic("muxer uninitialized")
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
