package tunutil

import (
	"context"
	"net"

	"github.com/xmx/aegis-common/tunnel/tunopen"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type DialMatcher interface {
	MatchDialer(network, address string) Dialer
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

func NewMatchDialer(fallback Dialer, matches ...DialMatcher) Dialer {
	return &matchDialer{
		matches:  matches,
		fallback: fallback,
	}
}

type matchDialer struct {
	matches  []DialMatcher
	fallback Dialer
}

func (md *matchDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	for _, m := range md.matches {
		if d := m.MatchDialer(network, address); d != nil {
			return d.DialContext(ctx, network, address)
		}
	}

	return md.fallback.DialContext(ctx, network, address)
}

func NewMuxDialer(mux tunopen.Muxer) Dialer {
	return &muxDialer{
		mux: mux,
	}
}

type muxDialer struct {
	mux tunopen.Muxer
}

func (m *muxDialer) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	return m.mux.Open(ctx)
}

func NewHostMatchDialer(host string, dial Dialer) DialMatcher {
	return &conditionDialer{
		match: func(_, address string) bool {
			addr, _, err := net.SplitHostPort(address)
			return err == nil && addr == host
		},
		dial: dial,
	}
}

type conditionDialer struct {
	match func(network, address string) bool
	dial  Dialer
}

func (cd *conditionDialer) MatchDialer(network, address string) Dialer {
	if cd.match(network, address) {
		return cd
	}

	return nil
}

func (cd *conditionDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return cd.dial.DialContext(ctx, network, address)
}
