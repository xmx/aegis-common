package logger

import (
	"context"
	"log/slog"
)

func NewSink(h slog.Handler, skips ...int) Sink {
	skip := 13
	if len(skips) > 0 {
		skip = skips[0]
	}
	han := Skip(h, skip)

	return Sink{
		log: slog.New(han),
	}
}

type Sink struct {
	log *slog.Logger
}

func (sk Sink) Info(level int, message string, keysAndValues ...any) {
	lvl := slog.LevelError
	if level == 1 {
		lvl = slog.LevelInfo
	} else if level == 2 {
		lvl = slog.LevelDebug
	}
	sk.log.Log(context.Background(), lvl, message, keysAndValues...)
}

func (sk Sink) Error(err error, message string, keysAndValues ...any) {
	kvs := []any{"error", err}
	kvs = append(kvs, keysAndValues...)

	sk.log.Error(message, kvs...)
}
