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
	Handle(tundial.Muxer) error
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
