package jsvm

import (
	"io"
	"sync"
	"sync/atomic"
)

type Writer interface {
	Write(p []byte) (n int, err error)
	Attach(w io.Writer) bool
	Detach(w io.Writer) bool
	Reset()
	Count() int
}

func newWriter() *dynamicWriter {
	dw := &dynamicWriter{writers: make(map[io.Writer]struct{})}
	dw.wrt.Store(new(multiWriter))

	return dw
}

type dynamicWriter struct {
	wrt     atomic.Pointer[multiWriter]
	mutex   sync.Mutex
	writers map[io.Writer]struct{}
}

func (dw *dynamicWriter) Attach(w io.Writer) bool {
	if w == nil {
		return false
	}
	if _, yes := w.(*dynamicWriter); yes {
		return false
	}
	if _, yes := w.(*multiWriter); yes {
		return false
	}

	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	if _, exists := dw.writers[w]; exists {
		return false
	}

	dw.writers[w] = struct{}{}
	dw.rewriter()

	return true
}

func (dw *dynamicWriter) Detach(w io.Writer) bool {
	if w == nil {
		return false
	}

	dw.mutex.Lock()
	defer dw.mutex.Unlock()

	if _, exists := dw.writers[w]; !exists {
		return false
	}

	delete(dw.writers, w)
	dw.rewriter()

	return true
}

func (dw *dynamicWriter) Reset() {
	dw.mutex.Lock()
	dw.writers = make(map[io.Writer]struct{}, 8)
	dw.rewriter()
	dw.mutex.Unlock()
}

func (dw *dynamicWriter) Count() int {
	wrt := dw.wrt.Load()
	return len(wrt.writers)
}

func (dw *dynamicWriter) rewriter() {
	writers := make([]io.Writer, 0, len(dw.writers))
	for w := range dw.writers {
		pw := &proxyWriter{dynamic: dw, writer: w}
		writers = append(writers, pw)
	}
	mw := &multiWriter{writers: writers}
	dw.wrt.Store(mw)
}

func (dw *dynamicWriter) Write(p []byte) (n int, err error) {
	wrt := dw.wrt.Load()
	return wrt.Write(p)
}

type proxyWriter struct {
	dynamic  *dynamicWriter
	failures int
	writer   io.Writer
}

func (pw *proxyWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err == nil {
		pw.failures = 0
	} else {
		if pw.failures++; pw.failures >= 3 {
			pw.dynamic.Detach(pw.writer)
		}
	}

	return n, err
}

type multiWriter struct {
	writers []io.Writer
}

func (t *multiWriter) Write(p []byte) (int, error) {
	for _, w := range t.writers {
		_, _ = w.Write(p)
	}
	return len(p), nil
}
