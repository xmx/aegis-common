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

// NewFirstMatchDialer returns a Dialer that tries each dialer in order
// and returns the first successful connection or error.
// A dialer returning (nil, nil) is treated as "no match" and the next dialer is tried.
// The back dialer is used as a fallback if no dialer matches.
func NewFirstMatchDialer(dials []Dialer, back Dialer) Dialer {
	return &firstMatchDialer{dials: dials, back: back}
}

type firstMatchDialer struct {
	dials []Dialer
	back  Dialer
}

func (fmd *firstMatchDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	for _, dial := range fmd.dials {
		if conn, err := dial.DialContext(ctx, network, address); conn != nil || err != nil {
			return conn, err
		}
	}

	return fmd.back.DialContext(ctx, network, address)
}

// NewMatchHostDialer creates a Dialer that only handles connections
// to the specified host. It uses the provided muxconn.Muxer to open a connection
// when the host matches.
//
// DialContext behavior:
//   - If the host part of the given address does not match mhd.host, it returns (nil, nil),
//     indicating "no match" so that a higher-level dialer can try other options.
//   - If the host matches, it opens a connection through mhd.mux.
//   - Any error from mux.Open is returned directly.
func NewMatchHostDialer(host string, mux muxconn.Muxer) Dialer {
	return &matchHostDialer{host: host, mux: mux}
}

type matchHostDialer struct {
	mux  muxconn.Muxer
	host string
}

func (mhd *matchHostDialer) DialContext(ctx context.Context, _, address string) (net.Conn, error) {
	if host, _, err := net.SplitHostPort(address); err != nil {
		return nil, nil
	} else if host != mhd.host {
		return nil, nil
	}

	return mhd.mux.Open(ctx)
}
