package jsvm

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/grafana/sobek"
)

var ErrRuntimeClosed = errors.New("jsvm: runtime closed")

type Engineer interface {
	Context() context.Context
	Runtime() *sobek.Runtime
	Compile(name, code string, strict bool) (*sobek.Program, error)
	RunScript(name, code string) (sobek.Value, error)
	RunProgram(pgm *sobek.Program) (sobek.Value, error)
	Output() (stdout, stderr Writer)
	Require() Requirer
	Defer() Defer

	// Kill 结束虚拟机
	Kill(v any) error
}

func NewVM(parent context.Context) Engineer {
	if parent == nil {
		parent = context.Background()
	}

	vm := sobek.New()
	vm.SetFieldNameMapper(tagMapper("json"))
	ctx, cancel := context.WithCancelCause(parent)

	svm := &sobekVM{
		vm:     vm,
		stdout: newWriter(),
		stderr: newWriter(),
		ctx:    ctx,
		cancel: cancel,
	}
	svm.deferment = newDeferment(svm)
	svm.require = newRequire(svm)

	return svm
}

type sobekVM struct {
	vm        *sobek.Runtime
	require   *sobekRequire
	stdout    Writer
	stderr    Writer
	closed    atomic.Bool
	deferment *deferment
	ctx       context.Context
	cancel    context.CancelCauseFunc
}

func (svm *sobekVM) Context() context.Context {
	return svm.ctx
}

func (svm *sobekVM) Runtime() *sobek.Runtime {
	return svm.vm
}

func (svm *sobekVM) Compile(name, code string, strict bool) (*sobek.Program, error) {
	cjs, err := Transform(name, code)
	if err != nil {
		return nil, err
	}

	return sobek.Compile(name, cjs, strict)
}

func (svm *sobekVM) RunScript(name, code string) (sobek.Value, error) {
	pgm, err := svm.Compile(name, code, false)
	if err != nil {
		return nil, err
	}

	return svm.RunProgram(pgm)
}

func (svm *sobekVM) RunProgram(pgm *sobek.Program) (sobek.Value, error) {
	if err := svm.closedError(); err != nil {
		return nil, err
	}

	return svm.vm.RunProgram(pgm)
}

func (svm *sobekVM) Output() (Writer, Writer) {
	return svm.stdout, svm.stderr
}

func (svm *sobekVM) Require() Requirer {
	return svm.require
}

func (svm *sobekVM) Defer() Defer {
	return svm.deferment
}

func (svm *sobekVM) Kill(v any) error {
	if err := svm.closedError(); err != nil {
		return err
	}

	var kerr error
	if ve, ok := v.(error); ok {
		kerr = ve
	} else {
		kerr = fmt.Errorf("killed: %v", v)
	}

	err := svm.deferment.call()
	svm.vm.Interrupt(kerr)
	svm.cancel(kerr)

	return err
}

func (svm *sobekVM) closedError() error {
	if svm.closed.Load() {
		return ErrRuntimeClosed
	}

	return nil
}
