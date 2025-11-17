package jsmod

import (
	"context"
	"net"
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

func (mod *httpModule) Preload(eng jsvm.Engineer) (string, any, bool) {
	mod.eng = eng
	vals := map[string]any{
		"createServer": mod.createServer,
		"newServeMux":  http.NewServeMux,
	}

	return "net/http", vals, true
}

func (mod *httpModule) createServer(opt httpServerOptions) (*sobek.Object, error) {
	vm := mod.eng.Runtime()
	parent := mod.eng.Context()

	ctx, cancel := context.WithCancelCause(parent)
	hts := &httpServer{eng: mod.eng, opt: opt, ctx: ctx, cancel: cancel}

	ret := vm.NewObject()
	if err := ret.Set("listen", hts.listen); err != nil {
		return nil, err
	}
	srv := &http.Server{
		Addr:              opt.Addr,
		Handler:           hts,
		ReadTimeout:       opt.ReadTimeout,
		ReadHeaderTimeout: opt.ReadHeaderTimeout,
		WriteTimeout:      opt.WriteTimeout,
		IdleTimeout:       opt.IdleTimeout,
		MaxHeaderBytes:    opt.MaxHeaderBytes,
	}
	hts.srv = srv

	return ret, nil
}

type httpServer struct {
	eng     jsvm.Engineer
	opt     httpServerOptions
	srv     *http.Server
	finalID uint64
	mutex   sync.Mutex
	closed  atomic.Bool
	ctx     context.Context
	cancel  context.CancelCauseFunc
}

func (hts *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hts.mutex.Lock()
	defer hts.mutex.Unlock()

	if h := hts.opt.Handler; h != nil {
		h.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (hts *httpServer) listen(on ...func()) (*sobek.Object, error) {
	vm := hts.eng.Runtime()
	ret := vm.NewObject()
	if err := ret.Set("wait", hts.wait); err != nil {
		return nil, err
	}

	addr := hts.opt.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	if len(on) != 0 && on[0] != nil {
		on[0]()
	}

	parent := hts.eng.Context()
	closeFn := func() { _ = hts.close(context.Canceled) }
	final := hts.eng.Finalizer()
	hts.finalID = final.Add(closeFn)
	context.AfterFunc(parent, closeFn)

	go func() {
		err1 := hts.srv.Serve(ln)
		_ = hts.close(err1)
	}()

	return ret, nil
}

func (hts *httpServer) Close() error {
	return hts.close(context.Canceled)
}

func (hts *httpServer) wait() error {
	<-hts.ctx.Done()
	return hts.ctx.Err()
}

func (hts *httpServer) close(cause error) error {
	if !hts.closed.CompareAndSwap(false, true) {
		return http.ErrServerClosed
	}

	if fid := hts.finalID; fid != 0 {
		hts.finalID = 0
		final := hts.eng.Finalizer()
		final.Del(fid)
	}
	hts.cancel(cause)

	return hts.srv.Close()
}
