package tundial

import (
	"context"
	"net"

	"github.com/xmx/aegis-common/tunnel/tunopen"
)

// ContextDialer defines an interface for dialing network connections
// with optional context support.
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

// NewFirstMatchDialer returns a ContextDialer that tries each dialer in order
// and returns the first successful connection or error.
// A dialer returning (nil, nil) is treated as "no match" and the next dialer is tried.
// The back dialer is used as a fallback if no dialer matches.
func NewFirstMatchDialer(dials []ContextDialer, back ContextDialer) ContextDialer {
	return &firstMatchDialer{dials: dials, back: back}
}

type firstMatchDialer struct {
	dials []ContextDialer
	back  ContextDialer
}

func (fmd *firstMatchDialer) Dial(network, address string) (net.Conn, error) {
	return fmd.DialContext(context.Background(), network, address)
}

func (fmd *firstMatchDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	for _, dial := range fmd.dials {
		if conn, err := dial.DialContext(ctx, network, address); conn != nil || err != nil {
			return conn, err
		}
	}

	return fmd.back.DialContext(ctx, network, address)
}

// NewMatchHostDialer creates a ContextDialer that only handles connections
// to the specified host. It uses the provided tunopen.Muxer to open a connection
// when the host matches.
//
// DialContext behavior:
//   - If the host part of the given address does not match mhd.host, it returns (nil, nil),
//     indicating "no match" so that a higher-level dialer can try other options.
//   - If the host matches, it opens a connection through mhd.mux.
//   - Any error from mux.Open is returned directly.
func NewMatchHostDialer(host string, mux tunopen.Muxer) ContextDialer {
	return &matchHostDialer{host: host, mux: mux}
}

type matchHostDialer struct {
	mux  tunopen.Muxer
	host string
}

func (mhd *matchHostDialer) Dial(network, address string) (net.Conn, error) {
	return mhd.DialContext(context.Background(), network, address)
}

func (mhd *matchHostDialer) DialContext(ctx context.Context, _, address string) (net.Conn, error) {
	if host, _, err := net.SplitHostPort(address); err != nil {
		return nil, nil
	} else if host != mhd.host {
		return nil, nil
	}

	return mhd.mux.Open(ctx)
}
