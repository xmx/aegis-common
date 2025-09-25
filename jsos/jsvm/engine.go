package jsvm

import (
	"context"

	"github.com/grafana/sobek"
)

type Engineer interface {
	Runtime() *sobek.Runtime
	Compile(name, code string, strict bool) (*sobek.Program, error)
	RunScript(name, code string) (sobek.Value, error)
	RunProgram(pgm *sobek.Program) (sobek.Value, error)
	Finalizer() Finalizer
	Require() Requirer
	Output() (stdout, stderr Writer)
	Kill(cause any)
	Context() context.Context
}

func New(parent context.Context) Engineer {
	vm := sobek.New()
	vm.SetFieldNameMapper(newJSONTagName())

	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)

	svm := &sobekVM{
		vm:        vm,
		stdout:    newWriter(),
		stderr:    newWriter(),
		finalizer: newFinalizer(),
		ctx:       ctx,
		cancel:    cancel,
	}
	require := injectRequire(svm)
	svm.require = require
	context.AfterFunc(ctx, func() {
		svm.interrupt(context.Canceled)
	})

	return svm
}

type sobekVM struct {
	vm        *sobek.Runtime
	stdout    *dynamicWriter
	stderr    *dynamicWriter
	finalizer *finalizer
	require   *sobekRequire
	ctx       context.Context
	cancel    context.CancelFunc
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
	return svm.vm.RunProgram(pgm)
}

func (svm *sobekVM) Finalizer() Finalizer {
	return svm.finalizer
}

func (svm *sobekVM) Require() Requirer {
	return svm.require
}

func (svm *sobekVM) Output() (Writer, Writer) {
	return svm.stdout, svm.stderr
}

func (svm *sobekVM) Kill(cause any) {
	svm.interrupt(cause)
	svm.cancel()
}

func (svm *sobekVM) interrupt(cause any) {
	svm.finalizer.finalize()
	svm.vm.Interrupt(cause)
}

func (svm *sobekVM) Context() context.Context {
	return svm.ctx
}
