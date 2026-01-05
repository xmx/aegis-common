//go:build !go1.26

package logger

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
)

func NewMultiHandler(hs ...slog.Handler) Handler {
	mh := new(multiHandler)
	mh.Replace(hs...)

	return mh
}

type multiHandler struct {
	mtx sync.Mutex
	out map[slog.Handler]struct{}
	ptr atomic.Pointer[slogMultiHandler]
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.load().Enabled(ctx, level)
}

func (h *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.load().Handle(ctx, record)
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h.load().WithAttrs(attrs)
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	return h.load().WithGroup(name)
}

func (h *multiHandler) Attach(hs ...slog.Handler) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if h.out == nil {
		h.out = make(map[slog.Handler]struct{}, 8)
	}
	for _, mh := range hs {
		if mh != nil {
			h.out[mh] = struct{}{}
		}
	}

	h.replace(h.out)
}

func (h *multiHandler) Detach(hs ...slog.Handler) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	for _, mh := range hs {
		delete(h.out, mh)
	}

	h.replace(h.out)
}

func (h *multiHandler) Replace(hs ...slog.Handler) {
	out := make(map[slog.Handler]struct{}, len(hs))
	arr := make([]slog.Handler, 0, len(hs))
	for _, mh := range hs {
		if mh == nil {
			continue
		}
		if _, exists := out[mh]; exists {
			continue
		}

		out[mh] = struct{}{}
		arr = append(arr, mh)
	}

	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.out = out
	mh := newMultiHandler(hs...)
	h.ptr.Store(mh)
}

func (h *multiHandler) load() *slogMultiHandler {
	if mh := h.ptr.Load(); mh != nil {
		return mh
	}
	return newMultiHandler()
}

func (h *multiHandler) replace(out map[slog.Handler]struct{}) {
	hs := make([]slog.Handler, 0, len(out))
	for k := range out {
		hs = append(hs, k)
	}

	h.out = out
	mh := newMultiHandler(hs...)
	h.ptr.Store(mh)
}

// slogMultiHandler is a [slog.Handler] that invokes all the given Handlers.
// Its Enable method reports whether any of the handlers' Enabled methods return true.
// Its Handle, WithAttr and WithGroup methods call the corresponding method on each of the enabled handlers.
type slogMultiHandler struct {
	multi []slog.Handler
}

func (h *slogMultiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.multi {
		if h.multi[i].Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (h *slogMultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for i := range h.multi {
		if h.multi[i].Enabled(ctx, r.Level) {
			if err := h.multi[i].Handle(ctx, r.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (h *slogMultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.multi))
	for i := range h.multi {
		handlers = append(handlers, h.multi[i].WithAttrs(attrs))
	}
	return &slogMultiHandler{multi: handlers}
}

func (h *slogMultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.multi))
	for i := range h.multi {
		handlers = append(handlers, h.multi[i].WithGroup(name))
	}
	return &slogMultiHandler{multi: handlers}
}

// newMultiHandler creates a [slogMultiHandler] with the given Handlers.
func newMultiHandler(handlers ...slog.Handler) *slogMultiHandler {
	h := make([]slog.Handler, len(handlers))
	copy(h, handlers)
	return &slogMultiHandler{multi: h}
}
