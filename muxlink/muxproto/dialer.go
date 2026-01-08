package muxproto

import (
	"context"
	"net"

	"github.com/xmx/aegis-common/muxlink/muxconn"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

func NewMUXDialer(mux muxconn.Muxer) Dialer {
	return &muxDialer{mux: mux}
}

type muxDialer struct {
	mux muxconn.Muxer
}

func (m muxDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	return m.mux.Open(ctx)
}
