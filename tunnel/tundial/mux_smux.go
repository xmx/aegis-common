package tundial

import (
	"context"
	"net"

	"github.com/xtaci/smux"
)

func NewSMUX(c net.Conn, cfg *smux.Config, isServer bool) (Muxer, error) {
	mux := &xtaciSMUX{
		laddr: c.LocalAddr(),
		raddr: c.RemoteAddr(),
	}
	var err error
	if isServer {
		mux.sess, err = smux.Server(c, cfg)
	} else {
		mux.sess, err = smux.Client(c, cfg)
	}
	if err != nil {
		return nil, err
	}

	return mux, nil
}

type xtaciSMUX struct {
	sess    *smux.Session
	laddr   net.Addr
	raddr   net.Addr
	traffic *trafficCounter
}

func (x *xtaciSMUX) Accept() (net.Conn, error) {
	stm, err := x.sess.AcceptStream()
	if err != nil {
		return nil, err
	}
	conn := x.makeConn(stm)

	return conn, nil
}

func (x *xtaciSMUX) Open(_ context.Context) (net.Conn, error) {
	stm, err := x.sess.OpenStream()
	if err != nil {
		return nil, err
	}
	conn := x.makeConn(stm)

	return conn, nil
}

func (x *xtaciSMUX) Close() error                 { return x.sess.Close() }
func (x *xtaciSMUX) Addr() net.Addr               { return x.laddr }
func (x *xtaciSMUX) RemoteAddr() net.Addr         { return x.raddr }
func (x *xtaciSMUX) Protocol() (string, string)   { return "tcp", "github.com/xtaci/smux" }
func (x *xtaciSMUX) Transferred() (rx, tx uint64) { return x.traffic.load() }

func (x *xtaciSMUX) makeConn(stm *smux.Stream) *smuxConn {
	return &smuxConn{stream: stm, traffic: x.traffic}
}
