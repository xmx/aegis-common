package muxconn

import (
	"context"
	"net"

	"github.com/xtaci/smux"
)

func NewSMUX(raw net.Conn, cfg *smux.Config, serverSide bool) (Muxer, error) {
	var err error
	var sess *smux.Session
	if serverSide {
		sess, err = smux.Server(raw, cfg)
	} else {
		sess, err = smux.Client(raw, cfg)
	}
	if err != nil {
		return nil, err
	}

	mux := &xtaciSMUX{
		sess:    sess,
		traffic: new(trafficStat),
	}

	return mux, nil
}

type xtaciSMUX struct {
	sess    *smux.Session
	traffic *trafficStat
}

func (x *xtaciSMUX) Accept() (net.Conn, error) {
	stm, err := x.sess.AcceptStream()
	if err != nil {
		return nil, err
	}

	return x.newConn(stm), nil
}

func (x *xtaciSMUX) Open(context.Context) (net.Conn, error) {
	stm, err := x.sess.OpenStream()
	if err != nil {
		return nil, err
	}

	return x.newConn(stm), nil
}

func (x *xtaciSMUX) Close() error               { return x.sess.Close() }
func (x *xtaciSMUX) Addr() net.Addr             { return x.sess.LocalAddr() }
func (x *xtaciSMUX) RemoteAddr() net.Addr       { return x.sess.RemoteAddr() }
func (x *xtaciSMUX) Protocol() (string, string) { return "tcp", "github.com/xtaci/smux" }
func (x *xtaciSMUX) Traffic() (uint64, uint64)  { return x.traffic.Load() }

func (x *xtaciSMUX) newConn(stm *smux.Stream) *xtaciConn {
	return &xtaciConn{stm: stm, mst: x}
}
