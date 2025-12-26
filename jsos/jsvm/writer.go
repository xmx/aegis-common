package jsvm

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

type Writer interface {
	io.Writer

	// Append 追加 io.Writer 并返回是否追加成功。
	Append(io.Writer) bool

	// Remove 移除 io.Writer 并返回是否移除成功。
	Remove(io.Writer) bool

	// Reset 移除所有的输出通道。
	Reset()
}

func newWriter() *outputWriter {
	return &outputWriter{
		idx: make(map[io.Writer]struct{}, 4),
	}
}

type outputWriter struct {
	mtx sync.Mutex
	idx map[io.Writer]struct{}
	out []io.Writer
	ptr atomic.Pointer[multiWriter]
}

func (ow *outputWriter) Write(p []byte) (int, error) {
	if w := ow.ptr.Load(); w != nil {
		return (*w).Write(p)
	}

	return len(p), nil
}

func (ow *outputWriter) Append(w io.Writer) bool {
	if w == nil {
		return false
	}

	ow.mtx.Lock()
	defer ow.mtx.Unlock()

	if _, exists := ow.idx[w]; exists {
		return false
	}

	ow.idx[w] = struct{}{}
	ow.out = append(ow.out, w)
	news := make(multiWriter, len(ow.out))
	copy(news, ow.out)
	ow.ptr.Store(&news)

	return true
}

func (ow *outputWriter) Remove(w io.Writer) bool {
	if w == nil {
		return false
	}

	ow.mtx.Lock()
	defer ow.mtx.Unlock()

	if _, exists := ow.idx[w]; !exists {
		return false
	}

	delete(ow.idx, w)
	outs := make([]io.Writer, 0, len(ow.idx))
	news := make(multiWriter, 0, len(ow.idx))
	for _, e := range ow.out { // 遍历数组，保留追加的顺序
		if e != w {
			outs = append(outs, e)
			news = append(news, e)
		}
	}
	ow.out = outs
	ow.ptr.Store(&news)

	return true
}

func (ow *outputWriter) Reset() {
	ow.mtx.Lock()
	defer ow.mtx.Unlock()

	ow.out = nil
	ow.idx = make(map[io.Writer]struct{})
	ow.ptr.Store(nil)
}

type multiWriter []io.Writer

func (mw multiWriter) Write(p []byte) (int, error) {
	n := len(p)
	var errs []error
	for _, w := range mw {
		if _, err := w.Write(p); err != nil {
			errs = append(errs, err)
		}
	}

	return n, errors.Join(errs...)
}
