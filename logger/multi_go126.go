//go:build go1.26

package logger

import (
	"context"
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
	ptr atomic.Pointer[slog.MultiHandler]
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
	mh := slog.NewMultiHandler(hs...)
	h.ptr.Store(mh)
}

func (h *multiHandler) load() *slog.MultiHandler {
	if mh := h.ptr.Load(); mh != nil {
		return mh
	}
	return slog.NewMultiHandler()
}

func (h *multiHandler) replace(out map[slog.Handler]struct{}) {
	hs := make([]slog.Handler, 0, len(out))
	for k := range out {
		hs = append(hs, k)
	}

	h.out = out
	mh := slog.NewMultiHandler(hs...)
	h.ptr.Store(mh)
}
