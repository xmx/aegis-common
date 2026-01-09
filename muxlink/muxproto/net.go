package muxproto

import (
	"io"
	"net"
	"sync"
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

func NewFlagClose(c io.Closer) *FlagClose {
	return &FlagClose{c: c}
}

type FlagClose struct {
	m sync.Mutex
	c io.Closer
	f bool
}

func (mc *FlagClose) Close() {
	mc.m.Lock()
	defer mc.m.Unlock()

	if !mc.f {
		_ = mc.c.Close()
		mc.f = true
	}
}

func (mc *FlagClose) Closed() bool {
	mc.m.Lock()
	defer mc.m.Unlock()

	return mc.f
}
