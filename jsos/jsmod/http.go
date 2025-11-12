package jsmod

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grafana/sobek"
	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewHTTP() jsvm.Module {
	return new(httpModule)
}

type httpServerOptions struct {
	Addr              string        `json:"addr"`
	Handler           http.Handler  `json:"handler"`
	ReadTimeout       time.Duration `json:"readTimeout"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout"`
	WriteTimeout      time.Duration `json:"writeTimeout"`
	IdleTimeout       time.Duration `json:"idleTimeout"`
	MaxHeaderBytes    int           `json:"maxHeaderBytes"`
}

type httpModule struct {
	eng jsvm.Engineer
}

func (hm *httpModule) Preload(eng jsvm.Engineer) (string, any, bool) {
	hm.eng = eng
	vals := map[string]any{
		"serve":       hm.serve,
		"newServeMux": http.NewServeMux,
	}

	return "net/http", vals, true
}

func (hm *httpModule) serve(opt httpServerOptions) (*sobek.Object, error) {
	handler := opt.Handler
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	parent := hm.eng.Context()
	ctx, cancel := context.WithCancel(parent)
	hs := &httpServer{eng: hm.eng, handler: handler, ctx: ctx, cancel: cancel}
	srv := &http.Server{
		Addr:              opt.Addr,
		Handler:           hs,
		ReadTimeout:       opt.ReadTimeout,
		ReadHeaderTimeout: opt.ReadHeaderTimeout,
		WriteTimeout:      opt.WriteTimeout,
		IdleTimeout:       opt.IdleTimeout,
		MaxHeaderBytes:    opt.MaxHeaderBytes,
	}
	final := hm.eng.Finalizer()
	hs.finalID = final.Add(hs.close)
	context.AfterFunc(ctx, hs.close)

	ret := hm.eng.Runtime().NewObject()
	if err := ret.Set("wait", hs.wait); err != nil {
		return nil, err
	}

	go func() {
		_ = srv.ListenAndServe()
		hs.close()
	}()

	return ret, nil
}

type httpServer struct {
	eng     jsvm.Engineer
	srv     *http.Server
	mutex   sync.Mutex
	handler http.Handler
	closed  atomic.Bool
	finalID uint64
	ctx     context.Context
	cancel  context.CancelFunc
}

func (hs *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hs.mutex.Lock()
	defer hs.mutex.Unlock()
	hs.handler.ServeHTTP(w, r)
}

func (hs *httpServer) wait() {
	<-hs.ctx.Done()
}

func (hs *httpServer) close() {
	_ = hs.Close()
}

func (hs *httpServer) Close() error {
	if !hs.closed.CompareAndSwap(false, true) {
		return nil
	}
	hs.cancel()

	if fid := hs.finalID; fid != 0 {
		hs.eng.Finalizer().Del(fid)
	}

	return hs.srv.Close()
}
