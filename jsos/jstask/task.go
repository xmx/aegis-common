package jstask

import (
	"sync/atomic"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

type Task struct {
	svm     jsvm.Engineer
	name    string
	code    string
	sha1sum string
	killed  atomic.Bool
}

func (tsk *Task) Info() (string, string, string) {
	return tsk.name, tsk.code, tsk.sha1sum
}

func (tsk *Task) Output() (jsvm.Writer, jsvm.Writer) {
	return tsk.svm.Output()
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
