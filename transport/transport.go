package transport

import (
	"context"
	"net"
	"net/http"
)

func NewHTTPTransport(ml MuxLoader, inner string) *http.Transport {
	dial := new(net.Dialer)
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if host, _, _ := net.SplitHostPort(addr); host == inner {
				if mux, ok := ml.LoadMux(); ok {
					return mux.Open(ctx)
				}
				return nil, &net.AddrError{
					Err:  "mux tunnel not initialized",
					Addr: addr,
				}
			}

			return dial.DialContext(ctx, network, addr)
		},
	}
}
