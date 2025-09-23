package logger

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
)

func NewHandler(hs ...slog.Handler) Handler {
	dh := &dynamicHandler{hds: make(map[slog.Handler]struct{}, 8)}
	dh.Attach(hs...)

	return dh
}

type Handler interface {
	slog.Handler
	Attach(hs ...slog.Handler)
	Detach(hs ...slog.Handler)
	Replace(hs ...slog.Handler)
}

type dynamicHandler struct {
	atm atomic.Value
	mtx sync.Mutex
	hds map[slog.Handler]struct{}
}

func (dh *dynamicHandler) Enabled(ctx context.Context, level slog.Level) bool {
	h := dh.atm.Load().(multiHandlers)
	return h.Enabled(ctx, level)
}

func (dh *dynamicHandler) Handle(ctx context.Context, record slog.Record) error {
	h := dh.atm.Load().(multiHandlers)
	return h.Handle(ctx, record)
}

func (dh *dynamicHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h := dh.atm.Load().(multiHandlers)
	return h.WithAttrs(attrs)
}

func (dh *dynamicHandler) WithGroup(name string) slog.Handler {
	h := dh.atm.Load().(multiHandlers)
	return h.WithGroup(name)
}

func (dh *dynamicHandler) Attach(hs ...slog.Handler) {
	dh.mtx.Lock()
	for _, h := range hs {
		if h == nil {
			continue
		}
		dh.hds[h] = struct{}{}
	}
	dh.reload()
	dh.mtx.Unlock()
}

func (dh *dynamicHandler) Detach(hs ...slog.Handler) {
	dh.mtx.Lock()
	for _, h := range hs {
		if h == nil {
			continue
		}
		delete(dh.hds, h)
	}
	dh.reload()
	dh.mtx.Unlock()
}

func (dh *dynamicHandler) Replace(hs ...slog.Handler) {
	dh.mtx.Lock()
	clear(dh.hds)
	for _, h := range hs {
		if h == nil {
			continue
		}
		dh.hds[h] = struct{}{}
	}
	dh.reload()
	dh.mtx.Unlock()
}

func (dh *dynamicHandler) reload() {
	mhs := make(multiHandlers, 0, len(dh.hds))
	for h := range dh.hds {
		mhs = append(mhs, h)
	}
	dh.atm.Store(mhs)
}

type multiHandlers []slog.Handler

func (mhs multiHandlers) Enabled(ctx context.Context, l slog.Level) bool {
	for _, h := range mhs {
		if h.Enabled(ctx, l) {
			return true
		}
	}

	return false
}

func (mhs multiHandlers) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, h := range mhs {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (mhs multiHandlers) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(mhs))
	for _, h := range mhs {
		handlers = append(handlers, h.WithAttrs(attrs))
	}

	return multiHandlers(handlers)
}

func (mhs multiHandlers) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(mhs))
	for _, h := range mhs {
		handlers = append(handlers, h.WithGroup(name))
	}

	return multiHandlers(handlers)
}
