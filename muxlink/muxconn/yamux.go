package muxconn

import (
	"context"
	"net"

	"github.com/hashicorp/yamux"
	"golang.org/x/time/rate"
)

func NewYaMUX(parent context.Context, conn net.Conn, cfg *yamux.Config, serverSide bool) (Muxer, error) {
	if parent == nil {
		parent = context.Background()
	}

	var err error
	mux := &yamuxSession{
		traffic: new(trafficStat),
		limiter: newUnlimit(),
		streams: new(streamStat),
	}

	if serverSide {
		mux.sess, err = yamux.Server(conn, cfg)
	} else {
		mux.sess, err = yamux.Client(conn, cfg)
	}
	if err != nil {
		return nil, err
	}

	return mux, nil
}

type yamuxSession struct {
	sess    *yamux.Session
	traffic *trafficStat
	limiter *rateLimiter // 读写限流器
	streams *streamStat
	parent  context.Context
}

func (m *yamuxSession) Accept() (net.Conn, error)              { return m.newConn(m.sess.AcceptStream()) }
func (m *yamuxSession) Close() error                           { return m.sess.Close() }
func (m *yamuxSession) Addr() net.Addr                         { return m.sess.LocalAddr() }
func (m *yamuxSession) Open(context.Context) (net.Conn, error) { return m.newConn(m.sess.OpenStream()) }
func (m *yamuxSession) RemoteAddr() net.Addr                   { return m.sess.RemoteAddr() }
func (m *yamuxSession) Library() (string, string)              { return "yamux", "github.com/hashicorp/yamux" }
func (m *yamuxSession) Traffic() (uint64, uint64)              { return m.traffic.Load() }
func (m *yamuxSession) Limit() rate.Limit                      { return m.limiter.Limit() }
func (m *yamuxSession) SetLimit(bps rate.Limit)                { m.limiter.SetLimit(bps) }
func (m *yamuxSession) NumStreams() (int64, int64)             { return m.streams.Load() }

func (m *yamuxSession) newConn(stm *yamux.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(m.parent)
	lrw := m.limiter.newReadWriter(ctx, stm)
	m.streams.incr()

	conn := &yamuxConn{
		master: m,
		stream: stm,
		limit:  lrw,
		cancel: cancel,
	}

	return conn, nil
}
