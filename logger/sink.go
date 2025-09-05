package logger

import (
	"context"
	"log/slog"
)

func NewSink(h slog.Handler, skip int) Sink {
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
