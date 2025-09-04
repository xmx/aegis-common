package transport

import (
	"context"
	"io"
	"net"

	"github.com/xtaci/smux"
	"golang.org/x/net/quic"
)

type Handler interface {
	Handle(Muxer) error
}

type Server interface {
	Serve(net.Listener) error
}

type Muxer interface {
	// Open 打开一个子流。
	Open(ctx context.Context) (net.Conn, error)

	// Accept 一个子流（双向流）。
	Accept() (net.Conn, error)

	// Addr returns the listener's network address.
	Addr() net.Addr

	// Close 关闭多路复用，此操作会中断所有的子流。
	Close() error

	Protocol() string

	RemoteAddr() net.Addr
}

type tcpMux struct {
	sess *smux.Session
}

func (tm *tcpMux) Open(_ context.Context) (net.Conn, error) {
	stm, err := tm.sess.OpenStream()
	if err != nil {
		return nil, err
	}

	return stm, nil
}

func (tm *tcpMux) Accept() (net.Conn, error) {
	stm, err := tm.sess.AcceptStream()
	if err != nil {
		return nil, err
	}

	return stm, nil
}

func (tm *tcpMux) Addr() net.Addr {
	return tm.sess.LocalAddr()
}

func (tm *tcpMux) Close() error {
	return tm.sess.Close()
}

func (tm *tcpMux) Protocol() string {
	return "tcp"
}

func (tm *tcpMux) RemoteAddr() net.Addr {
	return tm.sess.RemoteAddr()
}

type udpMux struct {
	qc    *quic.Conn
	end   *quic.Endpoint
	laddr net.Addr
	raddr net.Addr
}

func (um *udpMux) Open(ctx context.Context) (net.Conn, error) {
	stm, err := um.qc.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	conn := um.newConn(stm)

	return conn, nil
}

func (um *udpMux) Accept() (net.Conn, error) {
	stm, err := um.qc.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}
	conn := um.newConn(stm)

	return conn, nil
}

func (um *udpMux) Addr() net.Addr {
	return um.laddr
}

func (um *udpMux) Close() error {
	return um.qc.Close()
}

func (um *udpMux) Protocol() string {
	return "udp"
}

func (um *udpMux) RemoteAddr() net.Addr {
	return um.raddr
}

func (um *udpMux) newConn(stm *quic.Stream) *quicConn {
	return &quicConn{
		stm:   stm,
		laddr: um.laddr,
		raddr: um.raddr,
	}
}

func NewSMUX(rwc io.ReadWriteCloser, server bool) (Muxer, error) {
	var err error
	var sess *smux.Session
	if server {
		sess, err = smux.Server(rwc, nil)
	} else {
		sess, err = smux.Client(rwc, nil)
	}
	if err != nil {
		return nil, err
	}
	tm := &tcpMux{sess: sess}

	return tm, nil
}
