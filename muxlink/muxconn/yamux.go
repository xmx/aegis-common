package muxconn

import (
	"context"
	"net"

	"github.com/hashicorp/yamux"
	"golang.org/x/time/rate"
)

func NewYaMUX(conn net.Conn, cfg *yamux.Config, serverSide bool) (Muxer, error) {
	var err error
	mux := &yamuxMUX{
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

type yamuxMUX struct {
	sess    *yamux.Session
	traffic *trafficStat
	limiter *rateLimiter // 读写限流器
	streams *streamStat
}

func (m *yamuxMUX) Accept() (net.Conn, error)              { return m.newConn(m.sess.AcceptStream()) }
func (m *yamuxMUX) Close() error                           { return m.sess.Close() }
func (m *yamuxMUX) Addr() net.Addr                         { return m.sess.LocalAddr() }
func (m *yamuxMUX) Open(context.Context) (net.Conn, error) { return m.newConn(m.sess.OpenStream()) }
func (m *yamuxMUX) RemoteAddr() net.Addr                   { return m.sess.RemoteAddr() }
func (m *yamuxMUX) Library() (string, string)              { return "yamux", "github.com/hashicorp/yamux" }
func (m *yamuxMUX) Traffic() (uint64, uint64)              { return m.traffic.Load() }
func (m *yamuxMUX) Limit() rate.Limit                      { return m.limiter.Limit() }
func (m *yamuxMUX) SetLimit(bps rate.Limit)                { m.limiter.SetLimit(bps) }
func (m *yamuxMUX) NumStreams() (int64, int64)             { return m.streams.NumStreams() }

func (m *yamuxMUX) newConn(stm *yamux.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	lrw := m.limiter.newReadWriter(ctx, stm)
	m.streams.openOne()

	conn := &yamuxConn{
		stm:    stm,
		mst:    m,
		lrw:    lrw,
		cancel: cancel,
	}

	return conn, nil
}
