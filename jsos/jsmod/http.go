package jsmod

import (
	"context"
	"net/http"
	"sync"

	"github.com/grafana/sobek"
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
	obj := svm.Runtime().NewObject()
	_ = obj.Set("listenAndServe", s.listenAndServe)
	_ = obj.Set("canonicalHeaderKey", http.CanonicalHeaderKey)
	_ = obj.Set("notFoundHandler", http.NotFoundHandler)

	return "net/http", obj, true
}

func (s *stdHTTP) listenAndServe(addr string, h http.Handler) (sobek.Value, error) {

	rt := s.svm.Runtime()
	obj := rt.NewObject()

	return nil, nil
}

type stdHTTPServer struct {
	svm    *sobek.Runtime
	srv    *http.Server
	fid    uint64
	ctx    context.Context
	cancel context.CancelFunc
}

func (s *stdHTTPServer) close() error {
	return s.srv.Close()
}

func (s *stdHTTPServer) wait(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	select {
	case <-ctx.Done():
	case <-s.ctx.Done():
	}
}
