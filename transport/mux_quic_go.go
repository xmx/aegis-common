package transport

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
)

type quicGoMux struct {
	conn  *quic.Conn
	laddr net.Addr
	raddr net.Addr
}

func (qm *quicGoMux) Open(ctx context.Context) (net.Conn, error) {
	stm, err := qm.conn.OpenStreamSync(ctx)
	if err != nil {
		return nil, err
	}
	conn := qm.makeConn(stm)

	return conn, nil
}

func (qm *quicGoMux) Accept() (net.Conn, error) {
	stm, err := qm.conn.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}
	conn := qm.makeConn(stm)

	return conn, nil
}

func (qm *quicGoMux) Addr() net.Addr {
	return qm.laddr
}

func (qm *quicGoMux) Close() error {
	return qm.conn.CloseWithError(0, "")
}

func (qm *quicGoMux) Protocol() string {
	return "udp"
}

func (qm *quicGoMux) RemoteAddr() net.Addr {
	return qm.raddr
}

func (qm *quicGoMux) makeConn(stm *quic.Stream) net.Conn {
	return &quicGoConn{
		stm:   stm,
		laddr: qm.laddr,
		raddr: qm.raddr,
	}
}
