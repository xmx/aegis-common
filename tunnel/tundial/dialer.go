package tundial

import (
	"context"
	"net"

	"github.com/xmx/aegis-common/tunnel/tunopen"
)

type ContextDialer interface {
	Dial(network, address string) (net.Conn, error)
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

func FromMuxer(mux tunopen.Muxer) ContextDialer {
	return &muxDialer{mux: mux}
}

type muxDialer struct {
	mux tunopen.Muxer
}

func (md *muxDialer) Dial(_, _ string) (net.Conn, error) {
	return md.mux.Open(context.Background())
}

func (md *muxDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	return md.mux.Open(ctx)
}
