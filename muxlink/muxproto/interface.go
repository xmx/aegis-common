package muxproto

import (
	"context"

	"github.com/xmx/aegis-common/muxlink/muxconn"
)

type MUXAccepter interface {
	AcceptMUX(muxconn.Muxer)
}

type ConfigLoader[T any] interface {
	LoadConfig(ctx context.Context) (*T, error)
}

type ServerHooker struct {
}

type ClientHooker struct {
}
