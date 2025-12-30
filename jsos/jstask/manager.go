package jstask

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/grafana/sobek"
	"github.com/xmx/aegis-common/jsos/jsvm"
)

type Manager interface {
	// Exec 执行程序。
	Exec(parent context.Context, name, code string) (bool, sobek.Value, error)

	// Lookup 通过名字查询任务。
	Lookup(name string) *Task

	// Kill 结束任务。
	Kill(name string, v any) error

	Tasks() Tasks
}

type Options struct {
	Logger  *slog.Logger
	Stdout  []io.Writer
	Stderr  []io.Writer
	Module  []jsvm.Module
	Context context.Context
}

type Optioner interface {
	Options() Options
}

func New(option Optioner) Manager {
	return &managers{
		option: option,
	}
}

type managers struct {
	option Optioner
	mutex  sync.RWMutex
	tasks  map[string]*Task
}

func (mana *managers) Exec(parent context.Context, name, code string) (bool, sobek.Value, error) {
	var opt Options
	if option := mana.option; option != nil {
		opt = option.Options()
	}
	if parent == nil {
		parent = opt.Context
	}
	if parent == nil {
		parent = context.Background()
	}

	sum := sha1.Sum([]byte(code))
	sha1sum := hex.EncodeToString(sum[:])

	mana.mutex.RLock()
	last := mana.tasks[name]
	mana.mutex.RUnlock()
	if last != nil && sha1sum == last.sha1sum {
		return false, nil, nil
	}

	svm := jsvm.NewVM(parent, opt.Logger)
	stdout, stderr := svm.Output()
	for _, out := range opt.Stdout {
		stdout.Append(out)
	}
	for _, out := range opt.Stderr {
		stderr.Append(out)
	}
	require := svm.Require()
	require.Registers(opt.Module...)

	task := &Task{
		svm:     svm,
		name:    name,
		code:    code,
		sha1sum: sha1sum,
		startAt: time.Now(),
	}

	if !mana.replaceTask(task) {
		return false, nil, nil
	}
	defer func() {
		mana.mutex.Lock()
		if tsk := mana.tasks[name]; tsk != nil && tsk.sha1sum == sha1sum {
			delete(mana.tasks, name)
		}
		mana.mutex.Unlock()
	}()

	val, err := svm.RunScript(name, code)
	_ = task.kill(&TaskError{Name: name, Text: "任务运行结束"})

	return true, val, err
}

func (mana *managers) Lookup(name string) *Task {
	mana.mutex.RLock()
	task := mana.tasks[name]
	mana.mutex.RUnlock()

	return task
}

func (mana *managers) Kill(name string, v any) error {
	mana.mutex.RLock()
	tsk := mana.tasks[name]
	mana.mutex.RUnlock()
	if tsk == nil {
		return &TaskError{Name: name, Text: "任务不存在"}
	}

	return tsk.kill(v)
}

func (mana *managers) Tasks() Tasks {
	mana.mutex.RLock()
	defer mana.mutex.RUnlock()

	tasks := make(Tasks, 0, len(mana.tasks))
	for _, tsk := range mana.tasks {
		tasks = append(tasks, tsk)
	}

	return tasks
}

func (mana *managers) replaceTask(tsk *Task) bool {
	name := tsk.name
	mana.mutex.Lock()
	defer mana.mutex.Unlock()

	if last := mana.tasks[name]; last != nil {
		if tsk.sha1sum == last.sha1sum {
			return false
		}
		_ = last.kill(&TaskError{Name: name, Text: "任务运行"})
	}

	if mana.tasks == nil {
		mana.tasks = make(map[string]*Task, 16)
	}
	mana.tasks[name] = tsk

	return true
}
