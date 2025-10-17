package httpkit

import (
	"net/http"
	"sync/atomic"
)

type Handler interface {
	http.Handler
	Store(http.Handler)
}

func NewHandler() Handler {
	return &atomicHandler{}
}

type atomicHandler struct {
	ptr atomic.Pointer[http.Handler]
}

func (ah *atomicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h := ah.ptr.Load(); h != nil {
		(*h).ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func (ah *atomicHandler) Store(h http.Handler) {
	if h == nil {
		ah.ptr.Store(nil)
	} else {
		ah.ptr.Store(&h)
	}
}
