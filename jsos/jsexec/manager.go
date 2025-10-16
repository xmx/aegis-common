package jsexec

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/xmx/aegis-common/jsos/jsvm"
	"github.com/xmx/aegis-common/options"
)

type Manager interface {
	Exec(ctx context.Context, name, code string)
	Find(pid uint64) Tasker
	Tasks() []Tasker
}

func NewManager(opts ...options.Lister[option]) Manager {
	opt := options.Eval(opts...)
	return &taskManager{
		opt:   opt,
		tasks: make(map[uint64]*jsTask, 64),
	}
}

type taskManager struct {
	opt   option
	pid   atomic.Uint64
	mutex sync.RWMutex
	tasks map[uint64]*jsTask
}

func (m *taskManager) Exec(ctx context.Context, name, code string) {
	if ctx == nil {
		ctx = context.Background()
	}
	pid := m.pid.Add(1)
	eng := jsvm.New(ctx)
	stdout, stderr := eng.Output()
	stdout.Attach(m.opt.stdout)
	stderr.Attach(m.opt.stderr)
	require := eng.Require()
	require.Registers(m.opt.modules)

	task := &jsTask{
		pid:  pid,
		name: name,
		code: code,
		eng:  eng,
	}

	m.mutex.Lock()
	m.tasks[pid] = task
	m.mutex.Unlock()

	defer func() {
		m.mutex.Lock()
		delete(m.tasks, pid)
		m.mutex.Unlock()
	}()

	task.exec(name, code)
	task.Kill("manager killed")
}

func (m *taskManager) Find(pid uint64) Tasker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if task := m.tasks[pid]; task != nil {
		return task
	}

	return nil
}

func (m *taskManager) Tasks() []Tasker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tasks := make([]Tasker, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}
