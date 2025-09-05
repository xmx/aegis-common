package logger

import (
	"context"
	"log/slog"
	"runtime"
)

func Skip(h slog.Handler, n int) slog.Handler {
	if n == 0 {
		return h
	}

	return &skipHandler{h: h, n: n}
}

type skipHandler struct {
	h slog.Handler
	n int
}

func (s *skipHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return s.h.Enabled(ctx, level)
}

func (s *skipHandler) Handle(ctx context.Context, record slog.Record) error {
	var pcs [1]uintptr
	runtime.Callers(s.n, pcs[:])
	record.PC = pcs[0]

	return s.h.Handle(ctx, record)
}

func (s *skipHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return s.h.WithAttrs(attrs)
}

func (s *skipHandler) WithGroup(name string) slog.Handler {
	return s.h.WithGroup(name)
}
