package transport

import (
	"context"
	"net"
)

type Handler interface {
	Handle(Muxer) error
}

type Server interface {
	Serve(net.Listener) error
}

type Muxer interface {
	// Open 打开一个子流。
	Open(ctx context.Context) (net.Conn, error)

	// Accept 一个子流（双向流）。
	Accept() (net.Conn, error)

	// Addr returns the listener's network address.
	Addr() net.Addr

	// Close 关闭多路复用，此操作会中断所有的子流。
	Close() error

	// Protocol 底层协议：udp/tcp
	Protocol() string

	RemoteAddr() net.Addr
}
