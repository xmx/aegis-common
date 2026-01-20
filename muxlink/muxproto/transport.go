package muxproto

import (
	"log/slog"
	"net/http"
	"time"
)

func newHTTPTransport(d Dialer, log *slog.Logger) *httpTransport {
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
	const chrome143 = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"
	r.Header.Set("User-Agent", chrome143)

	res, err := ht.tran.RoundTrip(r)
	attrs := []any{"method", r.Method, "url", r.URL}
	if err != nil {
		attrs = append(attrs, "error", err)
		ht.log.Warn("rpc client 网络层请求出错", attrs...)
	} else {
		ht.log.Debug("rpc client 网络层请求成功", attrs...)
	}

	return res, err
}
