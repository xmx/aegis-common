package jsmod

import (
	"net"

	"github.com/xmx/jsos/jsvm"
)

func NewNet() jsvm.Module {
	return new(stdNet)
}

type stdNet struct{}

func (*stdNet) Preload(jsvm.Engineer) (string, any, bool) {
	vals := map[string]any{
		"splitHostPort": net.SplitHostPort,
		"joinHostPort":  net.JoinHostPort,
		"parseCIDR":     net.ParseCIDR,
		"parseIP":       net.ParseIP,
		"parseMAC":      net.ParseMAC,
	}

	return "net", vals, false
}
