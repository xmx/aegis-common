package jsmod

import (
	"net/http/httputil"
	"net/url"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewHTTPUtil() jsvm.Module {
	return new(httputilModule)
}

type httputilModule struct {
	eng jsvm.Engineer
}

func (mod *httputilModule) Preload(eng jsvm.Engineer) (string, any, bool) {
	mod.eng = eng
	vals := map[string]any{
		"newSingleHostReverseProxy": mod.newSingleHostReverseProxy,
	}

	return "net/http/httputil", vals, true
}

func (mod *httputilModule) newSingleHostReverseProxy(rawURL string) (*httputil.ReverseProxy, error) {
	pu, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(pu)

	return proxy, nil
}
