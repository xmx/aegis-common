package jsmod

import (
	"context"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewContext() jsvm.Module {
	return new(stdContext)
}

type stdContext struct {
	eng jsvm.Engineer
}

func (mod *stdContext) Preload(eng jsvm.Engineer) (string, any, bool) {
	mod.eng = eng
	vals := map[string]any{
		"background":   context.Background,
		"withCancel":   context.WithCancel,
		"withTimeout":  context.WithTimeout,
		"withValue":    context.WithValue,
		"withDeadline": context.WithDeadline,
	}

	return "context", vals, false
}
