package muxtool

import (
	"io"
	"net"
	"sync/atomic"
	"time"
)

func Outbound(laddr, raddr net.Addr) net.IP {
	if addr, ok := laddr.(*net.TCPAddr); ok {
		return addr.IP
	}

	dest := raddr.String()
	conn, err := net.DialTimeout("udp", dest, time.Second)
	if err != nil {
		return net.IPv4zero
	}
	_ = conn.Close()
	saddr := conn.LocalAddr()
	if addr, ok := saddr.(*net.UDPAddr); ok {
		return addr.IP
	}

	return net.IPv4zero
}

type FlagCloser interface {
	// Close 关闭。
	Close()

	// Closed 是否已经关闭。
	Closed() bool
}

func NewFlagCloser(c io.Closer) FlagCloser {
	return &flagCloser{c: c}
}

type flagCloser struct {
	f atomic.Bool
	c io.Closer
}

func (f *flagCloser) Close() {
	if f.f.CompareAndSwap(false, true) {
		_ = f.c.Close()
	}
}

func (f *flagCloser) Closed() bool {
	return f.f.Load()
}
