package jsmod

import (
	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewVariable[T any](modname string) *Variable[T] {
	return &Variable[T]{
		name: modname,
	}
}

type Variable[T any] struct {
	name string
	data T
}

func (mod *Variable[T]) Preload(jsvm.Engineer) (string, any, bool) {
	vals := map[string]any{
		"get": mod.Get,
		"set": mod.Set,
	}

	name := mod.name
	if name == "" {
		name = "variable"
	}

	return name, vals, true
}

func (mod *Variable[T]) Get() T  { return mod.data }
func (mod *Variable[T]) Set(v T) { mod.data = v }
