package muxproto

import (
	"github.com/xmx/aegis-common/muxlink/muxconn"
)

type MUXAccepter interface {
	AcceptMUX(muxconn.Muxer)
}

type ClientHooker struct {
}
