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

	// Store 覆盖更新底层 Muxer。
	//
	// 客户端 Open 成功后，返回的 Muxer 交给业务层代码持有，但是任何连接都存在掉线重连的问题，
	// 重连成功后，要让业务层代码无感的替换底层的连接通道。
	Store(mux Muxer) (old Muxer)
}

func MakeAtomic(m Muxer) AtomicMuxer {
	if am, yes := m.(AtomicMuxer); yes {
		return am
	}
	am := new(atomicMuxer)
	_ = am.Store(m)

	return am
}

type atomicMuxer struct {
	val atomic.Pointer[atomicMuxHolder]
}

func (a *atomicMuxer) Accept() (net.Conn, error)                  { return a.get().Accept() }
func (a *atomicMuxer) Close() error                               { return a.get().Close() }
func (a *atomicMuxer) Addr() net.Addr                             { return a.get().Addr() }
func (a *atomicMuxer) Open(ctx context.Context) (net.Conn, error) { return a.get().Open(ctx) }
func (a *atomicMuxer) RemoteAddr() net.Addr                       { return a.get().RemoteAddr() }
func (a *atomicMuxer) Protocol() (string, string)                 { return a.get().Protocol() }

func (a *atomicMuxer) Store(mux Muxer) Muxer {
	m := &atomicMuxHolder{m: mux}
	if old := a.val.Swap(m); old != nil {
		return old.get()
	}

	return nil
}

func (a *atomicMuxer) get() Muxer {
	m := a.val.Load()
	return m.get()
}

type atomicMuxHolder struct{ m Muxer }

func (m atomicMuxHolder) get() Muxer { return m.m }

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
