package options

import "reflect"

type Lister[T any] interface {
	List() []func(before T) (after T)
}

func Eval[T any](ls ...Lister[T]) T {
	var v T
	for _, opt := range ls {
		if opt == nil || reflect.ValueOf(opt).IsNil() {
			continue
		}

		for _, eval := range opt.List() {
			v = eval(v)
		}
	}

	return v
}
