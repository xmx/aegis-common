package transport

import (
	"context"
	"io"
	"net"

	"github.com/xtaci/smux"
)

func NewSMUX(rwc io.ReadWriteCloser, client bool) (Muxer, error) {
	var err error
	var sess *smux.Session
	if client {
		sess, err = smux.Client(rwc, nil)
	} else {
		sess, err = smux.Server(rwc, nil)
	}
	if err != nil {
		return nil, err
	}
	mux := &smuxMux{sess: sess}

	return mux, nil
}

type smuxMux struct {
	sess *smux.Session
}

func (s *smuxMux) Open(context.Context) (net.Conn, error) {
	return s.sess.OpenStream()
}

func (s *smuxMux) Accept() (net.Conn, error) {
	return s.sess.AcceptStream()
}

func (s *smuxMux) Addr() net.Addr {
	return s.sess.LocalAddr()
}

func (s *smuxMux) Close() error {
	return s.sess.Close()
}

func (s *smuxMux) Protocol() string {
	return "tcp"
}

func (s *smuxMux) RemoteAddr() net.Addr {
	return s.sess.RemoteAddr()
}
