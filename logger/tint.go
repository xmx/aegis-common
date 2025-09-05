package logger

import (
	"io"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

func NewTint(w io.Writer) slog.Handler {
	return tint.NewHandler(w, &tint.Options{
		AddSource:  true,
		Level:      slog.LevelDebug,
		TimeFormat: time.RFC3339,
	})
}
