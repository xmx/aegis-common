package jsvm

import (
	"os"

	"github.com/grafana/sobek"
)

type Module interface {
	Preload(eng Engineer) (modname string, modvalue any, override bool)
}

type Requirer interface {
	// Register 注册模块。
	Register(mod Module)
	Registers(mods []Module)
}

func injectRequire(eng *sobekVM) *sobekRequire {
	sr := &sobekRequire{
		engine:  eng,
		modules: make(map[string]sobek.Value, 16),
		sources: make(map[string]sobek.Value, 16),
	}
	rt := eng.Runtime()
	_ = rt.Set("require", sr.require)

	return sr
}

type sobekRequire struct {
	engine  *sobekVM
	modules map[string]sobek.Value
	sources map[string]sobek.Value
}

func (sr *sobekRequire) Register(mod Module) {
	sr.Registers([]Module{mod})
}

func (sr *sobekRequire) Registers(mods []Module) {
	eng := sr.engine
	for _, mod := range mods {
		name, value, override := mod.Preload(eng)
		sr.register(name, value, override)
	}
}

// register 注册模块并返回是否注册成功。
func (sr *sobekRequire) register(name string, mod any, override bool) bool {
	_, exists := sr.modules[name]
	if exists && !override {
		return false
	}

	rt := sr.engine.Runtime()
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

	rt := sr.engine.Runtime()
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
