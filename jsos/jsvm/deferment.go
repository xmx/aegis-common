package jsvm

import (
	"errors"
	"slices"
	"sync"
)

type Defer interface {
	// Append 向 defer 逻辑中添加一个函数。
	Append(func() error) uint64

	// Remove 移除函数。
	Remove(fid uint64) func() error
}

func newDeferment(svm *sobekVM) *deferment {
	return &deferment{
		svm: svm,
		idx: make(map[uint64]*deferEntry, 4),
	}
}

type deferment struct {
	svm *sobekVM
	mtx sync.Mutex
	num uint64
	idx map[uint64]*deferEntry
}

func (d *deferment) Append(f func() error) uint64 {
	if err := d.svm.closedError(); err != nil {
		return 0
	}

	d.mtx.Lock()
	defer d.mtx.Unlock()

	d.num++
	fid := d.num
	ent := &deferEntry{id: fid, fn: f}
	d.idx[fid] = ent

	return fid
}

func (d *deferment) Remove(fid uint64) func() error {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if ent := d.idx[fid]; ent != nil {
		return ent.fn
	}

	return nil
}

func (d *deferment) call() error {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	entries := make([]*deferEntry, 0, len(d.idx))
	for _, e := range d.idx {
		entries = append(entries, e)
	}
	d.idx = make(map[uint64]*deferEntry)
	// LIFO 执行 defer
	slices.SortFunc(entries, func(a, b *deferEntry) int { return int(a.id - b.id) })

	var errs []error
	for _, e := range entries {
		if err := e.call(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

type deferEntry struct {
	id uint64
	fn func() error
}

func (de *deferEntry) call() error {
	if fn := de.fn; fn != nil {
		return fn()
	}
	return nil
}
