package transport

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"slices"
	"time"

	"github.com/gorilla/websocket"
	quicgo "github.com/quic-go/quic-go"
	"golang.org/x/net/quic"
)

const (
	protocolTCP = "tcp"
	protocolUDP = "udp"
)

type DialConfig struct {
	Protocols       []string          // 连接协议
	QUICConfig      *quic.Config      // QUIC 配置
	WebSocketDialer *websocket.Dialer // websocket
	WebSocketPath   string            // 默认 /api/tunnel
	Parent          context.Context
}

func (dc *DialConfig) DialTimeout(addr string, timeout time.Duration) (Muxer, error) {
	if addr == "" {
		addr = "localhost:443"
	} else {
		if _, _, err := net.SplitHostPort(addr); err != nil {
			addr = net.JoinHostPort(addr, "443")
		}
	}
	modes := slices.DeleteFunc(dc.Protocols, func(s string) bool {
		return s != protocolUDP && s != protocolTCP
	})
	if len(modes) == 0 {
		modes = []string{protocolUDP, protocolTCP}
	}
	num := len(modes)
	pout := timeout / time.Duration(num)

	var errs []error
	for _, mode := range modes {
		switch mode {
		case protocolTCP:
			mux, err := dc.dialHTTP(addr, pout)
			if err != nil {
				errs = append(errs, err)
			} else {
				return mux, nil
			}
		default:
			mux, err := dc.dialQUICGo(addr, pout)
			if err != nil {
				errs = append(errs, err)
			} else {
				return mux, nil
			}
		}
	}

	return nil, errors.Join(errs...)
}

func (dc *DialConfig) dialQUIC(addr string, timeout time.Duration) (Muxer, error) {
	cfg := dc.quicConfig()
	endpoint, err := quic.Listen("udp", "", cfg)
	if err != nil {
		return nil, err
	}

	parent := dc.Parent
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	conn, err := endpoint.Dial(ctx, "udp", addr, cfg)
	if err != nil {
		cctx, ccancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = endpoint.Close(cctx)
		ccancel()
		return nil, err
	}

	mux := NewQUIC(parent, conn, endpoint)

	return mux, nil
}

func (dc *DialConfig) dialHTTP(addr string, timeout time.Duration) (Muxer, error) {
	reqURL := &url.URL{
		Scheme: "wss",
		Host:   addr,
		Path:   dc.WebSocketPath,
	}
	if reqURL.Path == "" {
		reqURL.Path = "/api/tunnel"
	}
	strURL := reqURL.String()

	parent := dc.Parent
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

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

func (dc *DialConfig) dialQUICGo(addr string, timeout time.Duration) (Muxer, error) {
	tlsConfig := dc.tlsConfig()
	if len(tlsConfig.NextProtos) == 0 {
		tlsConfig.NextProtos = []string{"aegis"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := quicgo.DialAddr(ctx, addr, tlsConfig, &quicgo.Config{
		KeepAlivePeriod: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	laddr, raddr := conn.LocalAddr(), conn.RemoteAddr()
	mux := &quicGoMux{
		conn:  conn,
		laddr: laddr,
		raddr: raddr,
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
		tlsConfig := dc.tlsConfig()
		if len(tlsConfig.NextProtos) == 0 {
			tlsConfig.NextProtos = []string{"aegis"}
		}
		qc.TLSConfig = tlsConfig
	}

	return qc
}

func (dc *DialConfig) tlsConfig() *tls.Config {
	return &tls.Config{
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
