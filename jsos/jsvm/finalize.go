package jsvm

import (
	"sort"
	"sync"
)

type Finalizer interface {
	Add(f func()) uint64
	Del(id uint64) func()
	finalize()
}

func newFinalizer() *finalizer {
	return &finalizer{
		funcs: make(map[uint64]*finalize, 8),
	}
}

type finalizer struct {
	mutex  sync.Mutex
	funcs  map[uint64]*finalize
	serial uint64
}

func (fz *finalizer) Add(f func()) uint64 {
	if f == nil {
		return 0
	}

	fz.mutex.Lock()
	if fz.funcs == nil {
		fz.funcs = make(map[uint64]*finalize, 8)
	}
	fz.serial++
	id := fz.serial
	fz.funcs[id] = &finalize{id: id, fn: f}
	fz.mutex.Unlock()

	return id
}

func (fz *finalizer) Del(id uint64) func() {
	fz.mutex.Lock()
	f := fz.funcs[id]
	if f != nil {
		delete(fz.funcs, id)
	}
	fz.mutex.Unlock()

	if f != nil {
		return f.fn
	}

	return nil
}

func (fz *finalizer) finalize() {
	fz.mutex.Lock()
	fns := make(finalizes, 0, len(fz.funcs))
	for _, f := range fz.funcs {
		fns = append(fns, f)
	}
	fz.funcs = nil
	fz.mutex.Unlock()

	fns.sort()
	for _, fn := range fns {
		fn.call()
	}
}

type finalize struct {
	id uint64
	fn func()
}

func (fz *finalize) call() {
	if fz.fn != nil {
		fz.fn()
	}
}

type finalizes []*finalize

// sort 按照 id 倒序排序。
func (fs finalizes) sort() {
	sort.Slice(fs, func(i, j int) bool {
		return fs[i].id > fs[j].id
	})
}
