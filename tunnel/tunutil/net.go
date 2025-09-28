package tunutil

import (
	"context"
	"net"

	"github.com/xmx/aegis-common/tunnel/tundial"
)

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

type Server interface {
	Serve(net.Listener) error
}

type Handler interface {
	Handle(tundial.Muxer) error
}
