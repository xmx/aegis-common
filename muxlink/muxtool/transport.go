package muxtool

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/xmx/aegis-common/muxlink/muxproto"
)

func newHTTPTransport(d muxproto.Dialer, log *slog.Logger) *httpTransport {
	return &httpTransport{
		log: log,
		tran: &http.Transport{
			DialContext:           d.DialContext,
			MaxConnsPerHost:       50,
			IdleConnTimeout:       3 * time.Minute,
			ResponseHeaderTimeout: time.Minute,
		},
	}
}

type httpTransport struct {
	log  *slog.Logger
	tran *http.Transport
}

func (ht *httpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	ht.log.Debug("rpc client 网络层请求成功", "method", r.Method, "url", r.URL)

	return ht.tran.RoundTrip(r)
}
