package tundial

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	quicgo "github.com/quic-go/quic-go"
	"golang.org/x/net/quic"
)

// Config 连接器配置。
type Config struct {
	// Protocols 连接协议，udp tcp。
	// 默认：按照 udp tcp 的顺序连接。
	Protocols []string `json:"protocols"`

	// Addresses 服务端地址。
	// 默认：localhost:443
	Addresses []string `json:"addresses"`

	// PerTimeout 每次连接时的超时时间。
	// 默认 30s。
	PerTimeout time.Duration `json:"per_timeout"`

	// WebSocketDialer tcp 模式下会通过 websocket 建立通道，
	// 其实自己实现的 HTTP Streamable 也可以，但是 websocket
	// 在有网关、反代的情况下，具有良好的通过性。
	WebSocketDialer *websocket.Dialer

	// WebSocketPath tcp 模式下 websocket 的请求路径。
	// 默认：/api/tunnel
	WebSocketPath string

	// QUICConfig 使用标准库 quic 连接时会用到。
	QUICConfig *quic.Config

	// QUICGoConfig 使用 quic-go 连接时会用到。
	QUICGoConfig *quicgo.Config

	// TLSConfig tcp/udp 模式下如果没有配置 TLSConfig 就会使用该配置。
	TLSConfig *tls.Config

	// Parent 总的 context，在 quic 模式下有用。
	Parent context.Context `json:"-"`
}

func Open(openCfg Config) (Muxer, error) {
	c := openCfg.preformat()
	var errs []error
	for _, proto := range c.Protocols {
		for _, addr := range c.Addresses {
			mux, err := c.open(proto, addr)
			if err == nil {
				return mux, nil
			}
			errs = append(errs, err)
		}
	}

	return nil, errors.Join(errs...)
}

func (c Config) open(proto, addr string) (Muxer, error) {
	if proto == "tcp" {
		return c.openHTTP(addr)
	} else {
		return c.openQUIC(addr)
	}
}

func (c Config) openHTTP(addr string) (Muxer, error) {
	reqURL := &url.URL{
		Scheme: "wss",
		Host:   addr,
		Path:   c.WebSocketPath,
	}
	strURL := reqURL.String()

	ctx, cancel := c.context()
	defer cancel()

	d := c.webSocketDialer()
	ws, _, err := d.DialContext(ctx, strURL, nil)
	if err != nil {
		return nil, err
	}
	conn := ws.NetConn()
	mux, err1 := NewSMUX(conn, nil, false)
	if err1 != nil {
		_ = conn.Close()
		return nil, err1
	}

	return mux, nil
}

func (c Config) openQUIC(addr string) (Muxer, error) {
	ctx, cancel := c.context()
	defer cancel()

	tlsCfg := c.tlsConfig(true)
	quicCfg := c.QUICGoConfig
	if quicCfg == nil {
		quicCfg = &quicgo.Config{
			KeepAlivePeriod: 10 * time.Second,
		}
	}
	conn, err := quicgo.DialAddr(ctx, addr, tlsCfg, quicCfg)
	if err != nil {
		return nil, err
	}
	mux := NewQUIC(c.Parent, conn)

	return mux, nil
}

// openStdQUIC 标准库的 quic 存在 bug，经常性的 context canceled。
func (c Config) openStdQUIC(addr string) (Muxer, error) {
	endpoint, err := quic.Listen("udp", "", nil)
	if err != nil {
		return nil, err
	}

	quicCfg := c.quicConfig()
	ctx, cancel := c.context()
	defer cancel()
	conn, err1 := endpoint.Dial(ctx, "udp", addr, quicCfg)
	if err1 != nil {
		tctx, tcancel := context.WithTimeout(c.Parent, time.Second)
		_ = endpoint.Close(tctx)
		tcancel()

		return nil, err1
	}
	mux := NewStdQUIC(c.Parent, endpoint, conn)

	return mux, nil
}

func (c Config) preformat() Config {
	if c.Parent == nil {
		c.Parent = context.Background()
	}
	if c.PerTimeout <= 0 {
		c.PerTimeout = 30 * time.Second
	}
	if c.WebSocketPath == "" {
		c.WebSocketPath = "/api/tunnel"
	}
	{
		uniq := make(map[string]struct{}, 4)
		used := make([]string, 0, 2)
		for _, proto := range c.Protocols {
			if _, ok := uniq[proto]; !ok {
				uniq[proto] = struct{}{}
				used = append(used, proto)
			}
		}
		if len(used) == 0 {
			used = append(used, "udp", "tcp")
		}
		c.Protocols = used
	}
	{
		size := len(c.Addresses)
		uniq := make(map[string]struct{}, size)
		used := make([]string, 0, size)
		for _, addr := range c.Addresses {
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
		c.Addresses = used
	}

	return c
}

func (c Config) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Parent, c.PerTimeout)
}

func (c Config) tlsConfig(isQUIC bool) *tls.Config {
	if cfg := c.TLSConfig; cfg != nil {
		return cfg
	}

	cfg := &tls.Config{InsecureSkipVerify: true}
	if isQUIC {
		cfg.NextProtos = []string{"aegis"}
		cfg.MinVersion = tls.VersionTLS13
	}

	return cfg
}

func (c Config) quicConfig() *quic.Config {
	cfg := c.QUICConfig
	if cfg == nil {
		cfg = &quic.Config{
			KeepAlivePeriod: 10 * time.Second,
		}
	}
	if cfg.TLSConfig == nil {
		cfg.TLSConfig = c.tlsConfig(true)
	}

	return cfg
}

func (c Config) quicGoConfig() *quicgo.Config {
	if cfg := c.QUICGoConfig; cfg != nil {
		return cfg
	}

	return &quicgo.Config{
		KeepAlivePeriod: 10 * time.Second,
	}
}

func (c Config) webSocketDialer() *websocket.Dialer {
	if d := c.WebSocketDialer; d != nil {
		return d
	}

	return &websocket.Dialer{
		TLSClientConfig:  c.tlsConfig(false),
		HandshakeTimeout: c.PerTimeout,
	}
}
