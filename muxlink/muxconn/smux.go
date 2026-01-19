package muxconn

import (
	"context"
	"net"

	"github.com/xtaci/smux"
	"golang.org/x/time/rate"
)

func NewSMUX(conn net.Conn, cfg *smux.Config, serverSide bool) (Muxer, error) {
	var err error
	mux := &xtaciSMUX{
		traffic: new(trafficStat),
		limiter: newUnlimit(),
	}

	if serverSide {
		mux.sess, err = smux.Server(conn, cfg)
	} else {
		mux.sess, err = smux.Client(conn, cfg)
	}
	if err != nil {
		return nil, err
	}

	return mux, nil
}

type xtaciSMUX struct {
	sess    *smux.Session
	traffic *trafficStat
	limiter *rateLimiter // 读写限流器
}

func (x *xtaciSMUX) Accept() (net.Conn, error)              { return x.newConn(x.sess.AcceptStream()) }
func (x *xtaciSMUX) Open(context.Context) (net.Conn, error) { return x.newConn(x.sess.OpenStream()) }
func (x *xtaciSMUX) Close() error                           { return x.sess.Close() }
func (x *xtaciSMUX) Addr() net.Addr                         { return x.sess.LocalAddr() }
func (x *xtaciSMUX) RemoteAddr() net.Addr                   { return x.sess.RemoteAddr() }
func (x *xtaciSMUX) Library() (string, string)              { return "smux", "github.com/xtaci/smux" }
func (x *xtaciSMUX) Traffic() (uint64, uint64)              { return x.traffic.Load() }
func (x *xtaciSMUX) Limit() rate.Limit                      { return x.limiter.Limit() }
func (x *xtaciSMUX) SetLimit(bps rate.Limit)                { x.limiter.SetLimit(bps) }

func (x *xtaciSMUX) newConn(stm *smux.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	parent := context.Background()
	lrw := x.limiter.newReadWriter(parent, stm)

	conn := &xtaciConn{
		stm: stm,
		mst: x,
		lrw: lrw,
	}

	return conn, nil
}
