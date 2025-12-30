package jstask

import (
	"cmp"
	"context"
	"slices"
	"sync/atomic"
	"time"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

type Task struct {
	svm     jsvm.Engineer
	name    string
	code    string
	sha1sum string
	startAt time.Time
	killed  atomic.Bool
}

func (tsk *Task) Info() (string, string, string) {
	return tsk.name, tsk.code, tsk.sha1sum
}

func (tsk *Task) Output() (jsvm.Writer, jsvm.Writer) {
	return tsk.svm.Output()
}

func (tsk *Task) StartAt() time.Time {
	return tsk.startAt
}

func (tsk *Task) Context() context.Context {
	return tsk.svm.Context()
}

func (tsk *Task) kill(v any) error {
	if !tsk.killed.CompareAndSwap(false, true) {
		return &TaskError{Name: tsk.name, Text: "任务已结束"}
	}

	return tsk.svm.Kill(v)
}

type TaskError struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

func (t *TaskError) Error() string {
	return t.Name + ": " + t.Text
}

type Tasks []*Task

func (tasks Tasks) Sort() {
	slices.SortFunc(tasks, func(a, b *Task) int {
		aat, bat := a.startAt, b.startAt
		if n := aat.Compare(bat); n != 0 {
			return n
		}
		return cmp.Compare(a.name, b.name)
	})
}
