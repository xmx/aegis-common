package logger

import (
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

func NewTint(w io.Writer, opt *slog.HandlerOptions) slog.Handler {
	tp := &tint.Options{TimeFormat: time.RFC3339}
	if opt != nil {
		tp.Level = opt.Level
		tp.AddSource = opt.AddSource
		tp.ReplaceAttr = opt.ReplaceAttr
	}

	return tint.NewHandler(w, tp)
}
