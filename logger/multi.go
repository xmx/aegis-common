package logger

import "log/slog"

type Handler interface {
	slog.Handler
	Attach(hs ...slog.Handler)
	Detach(hs ...slog.Handler)
	Replace(hs ...slog.Handler)
}
