package transport

import (
	"context"
	"crypto/tls"
	"net/netip"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/quic"
)

type DualDialer struct {
	WebSocketDialer *websocket.Dialer
	WebSocketPath   string
	QUICEndpoint    *quic.Endpoint
	QUICConfig      *quic.Config
}

func (d *DualDialer) DialContext(ctx context.Context, addr string) (Muxer, error) {
	if mux, _ := d.dialQUIC(ctx, addr); mux != nil {
		return mux, nil
	}

	return d.dialWebSocket(ctx, addr)
}

func (d *DualDialer) dialQUIC(ctx context.Context, addr string) (Muxer, error) {
	quicCfg := d.QUICConfig
	if quicCfg == nil {
		quicCfg = new(quic.Config)
	}
	if quicCfg.TLSConfig == nil {
		quicCfg.TLSConfig = d.tlsConfig()
	}

	endpoint := d.QUICEndpoint
	var tmpEndpoint *quic.Endpoint
	if endpoint == nil {
		end, err := quic.Listen("ucp", "", quicCfg)
		if err != nil {
			return nil, err
		}
		endpoint = end
		tmpEndpoint = end
	}

	sess, err := endpoint.Dial(ctx, "udp", addr, quicCfg)
	if err != nil {
		if tmpEndpoint != nil {
			_ = tmpEndpoint.Close(context.Background())
		}
		return nil, err
	}

	laddr := sess.LocalAddr()
	raddr := sess.RemoteAddr()
	mux := &udpMux{
		qc:    sess,
		end:   endpoint,
		laddr: &udpAddr{addr: laddr},
		raddr: &udpAddr{addr: raddr},
	}

	return mux, nil
}

func (d *DualDialer) dialWebSocket(ctx context.Context, addr string) (Muxer, error) {
	dial := d.WebSocketDialer
	if dial == nil {
		tlsCfg := d.tlsConfig()
		dial = &websocket.Dialer{
			TLSClientConfig:  tlsCfg,
			HandshakeTimeout: 30 * time.Second,
		}
	}

	reqPath := d.WebSocketPath
	if reqPath == "" {
		reqPath = "/api/channel"
	}
	rawURL := &url.URL{
		Scheme: "wss",
		Host:   addr,
		Path:   reqPath,
	}
	strURL := rawURL.String()

	ws, _, err := dial.DialContext(ctx, strURL, nil)
	if err != nil {
		return nil, err
	}
	conn := ws.NetConn()
	mux, err := NewSMUX(conn, false)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return mux, nil
}

func (d *DualDialer) tlsConfig() *tls.Config {
	if qc := d.QUICConfig; qc != nil {
		if cfg := qc.TLSConfig; cfg != nil {
			return cfg
		}
	}

	return &tls.Config{}
}

type udpAddr struct{ addr netip.AddrPort }

func (u udpAddr) Network() string { return "udp" }
func (u udpAddr) String() string  { return u.addr.String() }
