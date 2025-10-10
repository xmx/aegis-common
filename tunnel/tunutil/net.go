package tunutil

import (
	"context"
	"net"

	"github.com/xmx/aegis-common/tunnel/tundial"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type MatchedDialer interface {
	Dialer
	DialMatched(network, host, port string) bool
}

func DefaultDialer() Dialer {
	return &defaultDialer{}
}

type defaultDialer struct {
	d net.Dialer
}

func (d defaultDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.d.DialContext(ctx, network, address)
}

type Handler interface {
	Handle(tundial.Muxer)
}

func NewMatchDialer(fallback Dialer, matches ...MatchedDialer) Dialer {
	if fallback == nil {
		fallback = DefaultDialer()
	}

	return &matchedDialer{
		matches:  matches,
		fallback: fallback,
	}
}

type matchedDialer struct {
	matches  []MatchedDialer
	fallback Dialer
}

func (md *matchedDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	for _, m := range md.matches {
		if m.DialMatched(network, host, port) {
			return m.DialContext(ctx, network, address)
		}
	}

	return md.fallback.DialContext(ctx, network, address)
}

func NewHostMatch(host string, dial Dialer) MatchedDialer {
	return &hostMatch{
		host: host,
		dial: dial,
	}
}

type hostMatch struct {
	host string
	dial Dialer
}

func (h *hostMatch) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return h.dial.DialContext(ctx, network, address)
}

func (h *hostMatch) DialMatched(_, host, _ string) bool {
	return h.host == host
}

func NewTunnelDialer(mux tundial.Muxer) Dialer {
	return &tunnelDialer{
		mux: mux,
	}
}

type tunnelDialer struct {
	mux tundial.Muxer
}

func (t *tunnelDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	return t.mux.Open(ctx)
}
