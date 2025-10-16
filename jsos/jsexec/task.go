package jsexec

import (
	"sync/atomic"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

const (
	TaskInitialize uint32 = iota
	TaskRunning
	TaskStopped
	TaskPanicked
)

type Tasker interface {
	PID() uint64
	Status() uint32
	Name() string
	Kill(cause any)
	Error() error
	Engineer() jsvm.Engineer
}

type jsTask struct {
	pid    uint64
	name   string
	code   string
	eng    jsvm.Engineer
	err    error
	status atomic.Uint32
}

func (t *jsTask) PID() uint64 {
	return t.pid
}

func (t *jsTask) Status() uint32 {
	return t.status.Load()
}

func (t *jsTask) Name() string {
	return t.name
}

func (t *jsTask) Kill(cause any) {
	t.eng.Kill(cause)
}

func (t *jsTask) Error() error {
	return t.err
}

func (t *jsTask) Engineer() jsvm.Engineer {
	return t.eng
}

func (t *jsTask) exec(name, code string) {
	if !t.status.CompareAndSwap(TaskInitialize, TaskRunning) {
		return
	}
	defer func() {
		if v := recover(); v != nil {
			t.status.Store(TaskPanicked)
		} else {
			t.status.Store(TaskStopped)
		}
	}()

	_, err := t.eng.RunScript(name, code)
	t.err = err
}
