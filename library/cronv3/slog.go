package cronv3

import (
	"log/slog"

	"github.com/robfig/cron/v3"
)

func NewLog(l *slog.Logger) cron.Logger {
	return &cronLog{l: l}
}

type cronLog struct {
	l *slog.Logger
}

func (c *cronLog) Info(msg string, keysAndValues ...any) {
	c.l.Info(msg, keysAndValues...)
}

func (c *cronLog) Error(err error, msg string, keysAndValues ...any) {
	c.l.Error(msg, append([]any{"error", err}, keysAndValues...)...)
}
