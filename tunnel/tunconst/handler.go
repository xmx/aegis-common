package tunconst

import "github.com/xmx/aegis-common/tunnel/tunopen"

type Handler interface {
	Handle(tunopen.Muxer)
}

// NET -> RPC -> BIZ
