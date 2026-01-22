package muxproto

import (
	"context"
	"net"

	"github.com/xmx/aegis-common/muxlink/muxconn"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type MUXOpener interface {
	Host() string
	Open(ctx context.Context) (net.Conn, error)
}

func NewMUXOpener(mux muxconn.Muxer, host string) MUXOpener {
	return &muxOpen{mux: mux, host: host}
}

type muxOpen struct {
	mux  muxconn.Muxer
	host string
}

func (m *muxOpen) Host() string {
	return m.host
}

func (m *muxOpen) Open(ctx context.Context) (net.Conn, error) {
	return m.mux.Open(ctx)
}

func NewMixedDialer(mux MUXOpener, back Dialer) Dialer {
	return &mixedDialer{
		mux:  mux,
		back: back,
	}
}

type mixedDialer struct {
	mux  MUXOpener
	back Dialer
}

func (m *mixedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if m.mux != nil {
		host, _, err := net.SplitHostPort(address)
		if err == nil && host == m.mux.Host() {
			return m.mux.Open(ctx)
		}
	}

	if m.back != nil {
		return m.back.DialContext(ctx, network, address)
	}

	return nil, &net.OpError{
		Op:   "dial",
		Net:  network,
		Addr: &net.UnixAddr{Net: network, Name: address},
		Err:  net.UnknownNetworkError("没有找到任何拨号器"),
	}
}
