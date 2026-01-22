package muxconn

import (
	"context"
	"net"

	"github.com/xtaci/smux"
	"golang.org/x/time/rate"
)

func NewSMUX(parent context.Context, conn net.Conn, cfg *smux.Config, serverSide bool) (Muxer, error) {
	if parent == nil {
		parent = context.Background()
	}

	var err error
	mux := &smuxSession{
		traffic: new(trafficStat),
		limiter: newUnlimit(),
		streams: new(streamStat),
		parent:  parent,
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

type smuxSession struct {
	sess    *smux.Session
	traffic *trafficStat
	limiter *rateLimiter
	streams *streamStat
	parent  context.Context
}

func (m *smuxSession) Open(context.Context) (net.Conn, error) { return m.newConn(m.sess.OpenStream()) }
func (m *smuxSession) Accept() (net.Conn, error)              { return m.newConn(m.sess.AcceptStream()) }
func (m *smuxSession) Close() error                           { return m.sess.Close() }
func (m *smuxSession) Addr() net.Addr                         { return m.sess.LocalAddr() }
func (m *smuxSession) RemoteAddr() net.Addr                   { return m.sess.RemoteAddr() }
func (m *smuxSession) IsClosed() bool                         { return m.sess.IsClosed() }
func (m *smuxSession) Limit() rate.Limit                      { return m.limiter.Limit() }
func (m *smuxSession) SetLimit(bps rate.Limit)                { m.limiter.SetLimit(bps) }
func (m *smuxSession) NumStreams() (int64, int64)             { return m.streams.Load() }
func (m *smuxSession) Traffic() (uint64, uint64)              { return m.traffic.Load() }
func (m *smuxSession) Library() (string, string)              { return "smux", "github.com/xtaci/smux" }

func (m *smuxSession) newConn(stm *smux.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	m.streams.incr()
	ctx, cancel := context.WithCancelCause(m.parent)
	limit := m.limiter.newReadWriter(ctx, stm)

	conn := &smuxConn{
		parent: m,
		stream: stm,
		limit:  limit,
		cancel: cancel,
	}

	return conn, nil
}
