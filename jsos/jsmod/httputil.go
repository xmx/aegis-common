package jsmod

import (
	"net/http/httputil"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewHTTPUtil() jsvm.Module {
	return new(httputilModule)
}

type httputilModule struct {
	eng jsvm.Engineer
}

func (hum *httputilModule) Preload(eng jsvm.Engineer) (string, any, bool) {
	hum.eng = eng
	vals := map[string]any{
		"newSingleHostReverseProxy": httputil.NewSingleHostReverseProxy,
	}

	return "net/http/httputil", vals, true
}
