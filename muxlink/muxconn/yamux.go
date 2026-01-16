package muxconn

import (
	"context"
	"net"

	"github.com/hashicorp/yamux"
)

func NewYaMUX(conn net.Conn, cfg *yamux.Config, serverSide bool) (Muxer, error) {
	var err error
	mux := &yamuxMUX{traffic: new(trafficStat)}

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
}

func (m *yamuxMUX) Accept() (net.Conn, error)              { return m.newConn(m.sess.AcceptStream()) }
func (m *yamuxMUX) Close() error                           { return m.sess.Close() }
func (m *yamuxMUX) Addr() net.Addr                         { return m.sess.LocalAddr() }
func (m *yamuxMUX) Open(context.Context) (net.Conn, error) { return m.newConn(m.sess.OpenStream()) }
func (m *yamuxMUX) RemoteAddr() net.Addr                   { return m.sess.RemoteAddr() }
func (m *yamuxMUX) Library() (string, string)              { return "yamux", "github.com/hashicorp/yamux" }
func (m *yamuxMUX) Traffic() (uint64, uint64)              { return m.traffic.Load() }

func (m *yamuxMUX) newConn(stm *yamux.Stream, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}

	return &yamuxConn{stm: stm, mst: m}, nil

}
