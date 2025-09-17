package transport

import (
	"context"
	"net"
	"net/http"
	"net/url"
)

// server -> broker: <id>.broker.aegis.internal
// broker -> server: server.aegis.internal
//  broker -> agent: <id>.agent.aegis.internal
//  agent -> broker: broker.aegis.internal

const (
	BrokerHost = "broker.aegis.internal"
	ServerHost = "server.aegis.internal"
)

func NewAgentURL(id string, path string) *url.URL {
	return newURL(id+".aegis.internal", path)
}

func NewBrokerIDURL(id, path string) *url.URL {
	return newURL(id+"."+BrokerHost, path)
}

func NewBrokerURL(path string) *url.URL {
	return newURL(BrokerHost, path)
}

func NewServerURL(path string) *url.URL {
	return newURL(ServerHost, path)
}

func newURL(host, path string) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path,
	}
}

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
