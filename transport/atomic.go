package transport

import (
	"context"
	"net"
	"sync/atomic"
)

func NewAtomic(m Muxer) AtomicMuxer {
	a := new(atomicMuxer)
	a.Store(m)

	return a
}

type atomicMuxer struct {
	v atomic.Pointer[muxerHandle]
}

func (a *atomicMuxer) Open(ctx context.Context) (net.Conn, error) {
	if h, err := a.load(); err == nil {
		return h.Open(ctx)
	} else {
		return nil, err
	}
}

func (a *atomicMuxer) Accept() (net.Conn, error) {
	if h, err := a.load(); err == nil {
		return h.Accept()
	} else {
		return nil, err
	}
}

func (a *atomicMuxer) Addr() net.Addr {
	if h, _ := a.load(); h != nil {
		return h.Addr()
	}

	return &net.IPAddr{
		IP: net.IPv4zero,
	}
}

func (a *atomicMuxer) Close() error {
	if h, _ := a.load(); h != nil {
		return h.Close()
	}

	return nil
}

func (a *atomicMuxer) Protocol() string {
	if h, _ := a.load(); h != nil {
		return h.Protocol()
	}

	return "nil"
}

func (a *atomicMuxer) RemoteAddr() net.Addr {
	if h, _ := a.load(); h != nil {
		return h.RemoteAddr()
	}

	return &net.IPAddr{
		IP: net.IPv4zero,
	}
}

func (a *atomicMuxer) Store(m Muxer) {
	if m == nil {
		return
	}

	h := &muxerHandle{Muxer: m}
	a.v.Store(h)
}

func (a *atomicMuxer) load() (*muxerHandle, error) {
	if m := a.v.Load(); m != nil {
		return m, nil
	}

	return nil, net.UnknownNetworkError("atomic muxer not initialized")
}

type muxerHandle struct {
	Muxer
}
