package jsmod

import (
	"context"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewContext() jsvm.Module {
	return new(stdContext)
}

type stdContext struct{}

func (*stdContext) Preload(jsvm.Engineer) (string, any, bool) {
	vals := map[string]any{
		"background":   context.Background,
		"todo":         context.TODO,
		"withCancel":   context.WithCancel,
		"withTimeout":  context.WithTimeout,
		"withValue":    context.WithValue,
		"withDeadline": context.WithDeadline,
	}

	return "context", vals, false
}
