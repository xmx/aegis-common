package tunutil

import "github.com/xmx/aegis-common/tunnel/tundial"

type Handler interface {
	Handle(tundial.Muxer)
}
