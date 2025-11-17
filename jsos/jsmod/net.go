package jsmod

import (
	"net"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewNet() jsvm.Module {
	return new(stdNet)
}

type stdNet struct{}

func (sn *stdNet) Preload(jsvm.Engineer) (string, any, bool) {
	vals := map[string]any{
		"splitHostPort": net.SplitHostPort,
		"joinHostPort":  net.JoinHostPort,
		"parseCIDR":     net.ParseCIDR,
		"parseIP":       net.ParseIP,
		"parseMAC":      net.ParseMAC,
		"dial":          net.Dial,
		"dialTimeout":   net.DialTimeout,
	}

	return "net", vals, false
}
