package vterm

import (
	"os"
	"os/exec"
	"sync"
	"sync/atomic"

	"github.com/creack/pty"
)

func StartPTMX(cmd *exec.Cmd) (Typewriter, error) {
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	pt := &ptmxTTY{ptmx: ptmx, cmd: cmd}

	return pt, nil
}

type ptmxTTY struct {
	ptmx *os.File
	cmd  *exec.Cmd
	mtx  sync.RWMutex
	siz  atomic.Pointer[Winsize]
}

func (pt *ptmxTTY) Read(p []byte) (int, error) {
	return pt.ptmx.Read(p)
}

func (pt *ptmxTTY) Write(p []byte) (int, error) {
	return pt.ptmx.Write(p)
}

func (pt *ptmxTTY) Close() error {
	if proc := pt.cmd.Process; proc != nil {
		return proc.Kill()
	}

	return nil
}

func (pt *ptmxTTY) Size() (Winsize, error) {
	if sz := pt.siz.Load(); sz != nil {
		return *sz, nil
	}

	pt.mtx.RLock()
	defer pt.mtx.RUnlock()

	siz, err := pty.GetsizeFull(pt.ptmx)
	if err != nil {
		return Winsize{}, err
	}
	ret := Winsize{Cols: uint32(siz.Cols), Rows: uint32(siz.Rows)}
	pt.siz.Store(&ret)

	return ret, err
}

func (pt *ptmxTTY) Setsize(sz Winsize) error {
	dat := &pty.Winsize{
		Rows: uint16(sz.Rows),
		Cols: uint16(sz.Cols),
	}

	pt.mtx.Lock()
	defer pt.mtx.Unlock()

	if err := pty.Setsize(pt.ptmx, dat); err != nil {
		return err
	}
	pt.siz.Store(nil)

	return nil
}
