package httpkit

import (
	"net/http"
	"sync/atomic"
)

type Handler interface {
	http.Handler
	Store(http.Handler)
}

func NewAtomicHandler(h http.Handler) Handler {
	ah := new(atomicHolder)
	ah.Store(h)

	return ah
}

type atomicHolder struct {
	v atomic.Pointer[handlerHolder]
}

func (ah *atomicHolder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h := ah.v.Load(); h != nil {
		h.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (ah *atomicHolder) Store(h http.Handler) {
	if h != nil {
		hh := &handlerHolder{h: h}
		ah.v.Store(hh)
	}
}

type handlerHolder struct {
	h http.Handler
}

func (hh *handlerHolder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hh.h.ServeHTTP(w, r)
}
