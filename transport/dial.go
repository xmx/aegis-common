package transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/quic"
)

const (
	DialUDP DialMode = 1 << iota
	DialTCP
	DialAuto DialMode = DialUDP | DialTCP
)

type DialMode uint

func (m DialMode) String() string {
	if m&DialUDP != 0 {
		return "udp"
	}
	if m&DialTCP != 0 {
		return "tcp"
	}

	return "auto"
}

type DialConfig struct {
	DialMode        DialMode          // 连接模式
	QUICConfig      *quic.Config      // QUIC 配置
	WebSocketDialer *websocket.Dialer // websocket
	WebSocketPath   string            // 默认 /api/tunnel
	Parent          context.Context
}

func (dc *DialConfig) DialContext(ctx context.Context, addr string) (Muxer, error) {
	if addr == "" {
		addr = "127.0.0.1:443"
	} else {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			addr = net.JoinHostPort(addr, "443")
		}
	}

	mode := dc.DialMode
	if mode == DialUDP {
		return dc.dialQUIC(ctx, addr)
	} else if mode == DialTCP {
		return dc.dialHTTP(ctx, addr)
	}

	mux, err := dc.dialQUIC(ctx, addr)
	if err != nil {
		mux, err = dc.dialHTTP(ctx, addr)
	}

	return mux, err
}

func (dc *DialConfig) dialQUIC(ctx context.Context, addr string) (Muxer, error) {
	cfg := dc.quicConfig()
	endpoint, err := quic.Listen("udp", "", cfg)
	if err != nil {
		return nil, err
	}
	conn, err := endpoint.Dial(ctx, "udp", addr, cfg)
	if err != nil {
		_ = endpoint.Close(context.Background())
		return nil, err
	}

	parent := dc.Parent
	if parent == nil {
		parent = context.Background()
	}
	mux := NewQUIC(parent, conn, endpoint)

	return mux, nil
}

func (dc *DialConfig) dialHTTP(ctx context.Context, addr string) (Muxer, error) {
	reqURL := &url.URL{
		Scheme: "wss",
		Host:   addr,
		Path:   dc.WebSocketPath,
	}
	if reqURL.Path == "" {
		reqURL.Path = "/api/tunnel"
	}
	strURL := reqURL.String()

	wd := dc.webSocketDialer()
	ws, _, err := wd.DialContext(ctx, strURL, nil)
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

func (dc *DialConfig) quicConfig() *quic.Config {
	qc := dc.QUICConfig
	if qc == nil {
		qc = &quic.Config{
			KeepAlivePeriod: 10 * time.Second,
		}
	}
	if qc.TLSConfig == nil {
		qc.TLSConfig = dc.tlsConfig()
	}

	return qc
}

func (dc *DialConfig) tlsConfig() *tls.Config {
	return &tls.Config{
		NextProtos:         []string{"http/1.1", "aegis"},
		MinVersion:         tls.VersionTLS13, // quic 最低要求 TLS1.3
		InsecureSkipVerify: true,
	}
}

func (dc *DialConfig) webSocketDialer() *websocket.Dialer {
	if wd := dc.WebSocketDialer; wd != nil {
		return wd
	}

	return &websocket.Dialer{
		TLSClientConfig: dc.tlsConfig(),
	}
}
