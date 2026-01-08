package muxconn

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	quicgo "github.com/quic-go/quic-go"
	"golang.org/x/net/quic"
)

type DialConfig struct {
	Addresses []string

	// 连接协议
	// quic: golang.org/x/net/quic
	// quic-go: github.com/quic-go/quic-go
	// tcp: github.com/gorilla/websocket + github.com/xtaci/smux
	Protocols []string

	PerTimeout time.Duration

	QUICConfig *quic.Config

	QUICGoTLSConfig *tls.Config

	QUICGoConfig *quicgo.Config

	WebsocketDialer *websocket.Dialer

	WebsocketPath string

	Logger *slog.Logger

	Context context.Context
}

func Open(cfg DialConfig) (Muxer, error) {
	cfg = cfg.format()
	var errs []error
	for _, addr := range cfg.Addresses {
		for _, proto := range cfg.Protocols {
			attrs := []any{"proto", proto, "addr", addr}
			cfg.log().Debug("开始连接...", attrs...)
			mux, err := cfg.open(proto, addr)
			if err != nil {
				errs = append(errs, err)
				attrs = append(attrs, "error", err)
				cfg.log().Warn("连接失败", attrs...)
				continue
			}
			cfg.log().Info("连接成功", attrs...)

			return mux, nil
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	return nil, net.InvalidAddrError("empty address")
}

func (dc *DialConfig) open(proto, addr string) (Muxer, error) {
	switch proto {
	case "tcp":
		return dc.openTCP(addr)
	case "quic-go":
		return dc.openQUICgo(addr)
	default:
		return dc.openQUICx(addr)
	}
}

func (dc *DialConfig) openQUICx(addr string) (Muxer, error) {
	cfg := dc.quicConfig()
	endpoint, err := quic.Listen("udp", addr, cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := dc.perContext()
	defer cancel()

	conn, err := endpoint.Dial(ctx, "udp", addr, cfg)
	if err != nil {
		_ = endpoint.Close(ctx)
		return nil, err
	}
	mux := NewQUICx(dc.Context, endpoint, conn)

	return mux, nil
}

func (dc *DialConfig) openQUICgo(addr string) (Muxer, error) {
	tlsCfg := dc.QUICGoTLSConfig
	if tlsCfg == nil {
		tlsCfg = &tls.Config{
			NextProtos:         []string{"aegis"},
			InsecureSkipVerify: true,
		}
	}
	cfg := dc.QUICGoConfig
	if cfg == nil {
		cfg = &quicgo.Config{
			KeepAlivePeriod: 10 * time.Second,
		}
	}

	ctx, cancel := dc.perContext()
	defer cancel()

	conn, err := quicgo.DialAddr(ctx, addr, tlsCfg, cfg)
	if err != nil {
		return nil, err
	}
	mux := NewQUICgo(dc.Context, conn)

	return mux, nil
}

func (dc *DialConfig) openTCP(addr string) (Muxer, error) {
	reqURL := &url.URL{
		Scheme: "wss",
		Host:   addr,
		Path:   dc.WebsocketPath,
	}
	strURL := reqURL.String()

	ctx, cancel := dc.perContext()
	defer cancel()

	d := dc.websocketDialer()
	ws, _, err := d.DialContext(ctx, strURL, nil)
	if err != nil {
		return nil, err
	}
	conn := ws.NetConn()
	mux, err := NewSMUX(conn, nil, false)
	if err != nil {
		_ = ws.Close()
		return nil, err
	}

	return mux, nil
}

func (dc *DialConfig) perContext() (context.Context, context.CancelFunc) {
	if du := dc.PerTimeout; du > 0 {
		return context.WithTimeout(dc.Context, du)
	}

	return context.WithCancel(dc.Context)
}

func (dc *DialConfig) quicConfig() *quic.Config {
	if qc := dc.QUICConfig; qc != nil {
		return qc
	}

	return &quic.Config{
		TLSConfig: &tls.Config{
			NextProtos:         []string{"aegis"},
			InsecureSkipVerify: true,
		},
		HandshakeTimeout: dc.PerTimeout,
		KeepAlivePeriod:  10 * time.Second,
		QLogLogger:       dc.Logger,
	}
}

func (dc *DialConfig) websocketDialer() *websocket.Dialer {
	if d := dc.WebsocketDialer; d != nil {
		return d
	}

	return &websocket.Dialer{
		HandshakeTimeout: dc.PerTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func (dc *DialConfig) log() *slog.Logger {
	if l := dc.Logger; l != nil {
		return l
	}

	return slog.Default()
}

func (dc *DialConfig) format() DialConfig {
	ret := DialConfig{
		PerTimeout:      dc.PerTimeout,
		QUICConfig:      dc.QUICConfig,
		QUICGoTLSConfig: dc.QUICGoTLSConfig,
		QUICGoConfig:    dc.QUICGoConfig,
		WebsocketDialer: dc.WebsocketDialer,
		WebsocketPath:   dc.WebsocketPath,
		Logger:          dc.Logger,
		Context:         dc.Context,
	}

	{
		uniq := make(map[string]struct{}, 8)
		used := make([]string, 0, 8)
		for _, addr := range dc.Addresses {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				addr = net.JoinHostPort(addr, "443")
			}
			if _, ok := uniq[addr]; !ok {
				uniq[addr] = struct{}{}
				used = append(used, addr)
			}
		}
		if len(used) == 0 {
			used = append(used, "localhost:443")
		}
		ret.Addresses = used
	}
	{
		uniq := make(map[string]struct{}, 4)
		used := make([]string, 0, 2)
		for _, proto := range dc.Protocols {
			proto = strings.ToLower(proto)
			switch proto {
			case "quic", "quic-go", "tcp":
			default:
				continue
			}
			if _, ok := uniq[proto]; !ok {
				uniq[proto] = struct{}{}
				used = append(used, proto)
			}
		}
		if len(used) == 0 {
			used = append(used, "udp", "quic")
		}
		ret.Protocols = used
	}

	if ret.PerTimeout <= 0 {
		ret.PerTimeout = time.Minute
	}
	if ret.WebsocketPath == "" {
		ret.WebsocketPath = "/api/tunnel"
	}
	if ret.Context == nil {
		ret.Context = context.Background()
	}

	return ret
}
