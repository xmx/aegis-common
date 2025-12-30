package jsstd

import (
	"net/http"
	"sync"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewHTTP() jsvm.Module {
	return new(stdHTTP)
}

type stdHTTP struct {
	svm jsvm.Engineer
	mtx sync.Mutex
}

func (s *stdHTTP) Preload(svm jsvm.Engineer) (string, any, bool) {
	s.svm = svm
	vals := map[string]any{
		"listenAndServe":     s.listenAndServe,
		"canonicalHeaderKey": http.CanonicalHeaderKey,
		"notFoundHandler":    http.NotFoundHandler,
		"newServeMux":        http.NewServeMux,
	}

	return "net/http", vals, true
}

func (s *stdHTTP) listenAndServe(addr string, h http.Handler) error {
	if h == nil {
		h = http.NotFoundHandler()
	}

	sh := &safeHandler{han: h}
	srv := &http.Server{Addr: addr, Handler: sh}
	s.svm.Defer().Append(func() error {
		s.svm.Logger().Warn("net/http 服务停止了")
		return srv.Close()
	})

	return srv.ListenAndServe()
}

type safeHandler struct {
	mtx sync.Mutex
	han http.Handler
}

func (s *safeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.han.ServeHTTP(w, r)
}
