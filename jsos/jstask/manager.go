package jstask

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

type Manager interface {
	// Lookup 通过进程号查找进程信息。
	Lookup(pid uint64) Processor

	Exec(ctx context.Context, name, code string) error

	All() []Processor
}

func NewTaskManager(mods []jsvm.Module) Manager {
	return &taskManager{
		mods:  mods,
		procs: make(map[uint64]*process, 64),
	}
}

type taskManager struct {
	mods  []jsvm.Module
	pid   atomic.Uint64
	mutex sync.RWMutex
	procs map[uint64]*process
}

func (tm *taskManager) Lookup(pid uint64) Processor {
	tm.mutex.RLock()
	proc := tm.procs[pid]
	tm.mutex.RUnlock()
	if proc == nil {
		return nil
	}

	return proc
}

func (tm *taskManager) Exec(ctx context.Context, name, code string) error {
	proc := tm.createProcess(ctx, name)
	pid := proc.pid
	defer func() {
		proc.kill("process exited")

		tm.mutex.Lock()
		delete(tm.procs, pid)
		tm.mutex.Unlock()
	}()

	err := proc.run(name, code)
	proc.err = err

	return err
}

func (tm *taskManager) All() []Processor {
	tm.mutex.RLock()
	procs := make([]Processor, 0, len(tm.procs))
	for _, proc := range tm.procs {
		procs = append(procs, proc)
	}
	tm.mutex.RUnlock()

	return procs
}

func (tm *taskManager) createProcess(ctx context.Context, name string) *process {
	pid := tm.pid.Add(1)
	eng := jsvm.New(ctx)
	eng.Require().Registers(tm.mods)
	proc := &process{
		pid:  pid,
		eng:  eng,
		name: name,
	}

	tm.mutex.Lock()
	tm.procs[pid] = proc
	tm.mutex.Unlock()

	return proc
}
