package jsvm

import (
	"os"

	"github.com/grafana/sobek"
)

type Module interface {
	Preload(svm Engineer) (modname string, modvalue any, override bool)
}

type Requirer interface {
	// Registers 注册模块。
	Registers(mods ...Module)
}

func newRequire(svm *sobekVM) *sobekRequire {
	sr := &sobekRequire{
		svm:     svm,
		modules: make(map[string]sobek.Value, 16),
		sources: make(map[string]sobek.Value, 16),
	}
	rt := svm.Runtime()
	_ = rt.Set("require", sr.require)

	return sr
}

type sobekRequire struct {
	svm     *sobekVM
	modules map[string]sobek.Value
	sources map[string]sobek.Value
}

func (sr *sobekRequire) Registers(mods ...Module) {
	if sr.svm.closedError() != nil {
		return
	}

	svm := sr.svm
	for _, mod := range mods {
		name, value, override := mod.Preload(svm)
		sr.register(name, value, override)
	}
}

// register 注册模块并返回是否注册成功。
func (sr *sobekRequire) register(name string, mod any, override bool) bool {
	_, exists := sr.modules[name]
	if exists && !override {
		return false
	}

	rt := sr.svm.Runtime()
	value := rt.ToValue(mod)
	sr.modules[name] = value

	return true
}

func (sr *sobekRequire) require(call sobek.FunctionCall) sobek.Value {
	name := call.Argument(0).String()
	var err error
	if name != "" {
		val, exists := sr.loadBootstrap(name)
		if exists {
			return val
		}

		if val, exists, err = sr.loadApplication(name); err == nil && exists {
			return val
		}
	}

	rt := sr.svm.Runtime()
	if err != nil && !os.IsNotExist(err) {
		panic(rt.NewTypeError("cannot find module '%s': ", name, err.Error()))
	}

	panic(rt.NewTypeError("cannot find module '%s'", name))
}

func (sr *sobekRequire) loadBootstrap(name string) (sobek.Value, bool) {
	val, exists := sr.modules[name]
	return val, exists
}

func (sr *sobekRequire) loadApplication(name string) (sobek.Value, bool, error) {
	return nil, false, nil
}
