package logger

import (
	"bytes"
	"log"
	"log/slog"
)

func NewV1(l *slog.Logger) *log.Logger {
	w := &v1Writer{l: l}
	return log.New(w, "", 0)
}

type v1Writer struct {
	l *slog.Logger
}

func (v *v1Writer) Write(p []byte) (int, error) {
	n := len(p)
	s := bytes.TrimRight(p, "\n")
	v.l.Info(string(s))

	return n, nil
}
