package profile

import "github.com/xmx/aegis-common/jsos/jsvm"

type option struct {
	limit   int64
	modules []jsvm.Module // .js 配置文件用
	modname string        // .js 配置文件用
}

func NewOption() OptionBuilder {
	return OptionBuilder{}
}

type OptionBuilder struct {
	opts []func(option) option
}

func (ob OptionBuilder) List() []func(option) option {
	return ob.opts
}

func (ob OptionBuilder) Limit(n int64) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.limit = n
		return o
	})
	return ob
}

func (ob OptionBuilder) Modules(modules []jsvm.Module) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.modules = modules
		return o
	})
	return ob
}

func (ob OptionBuilder) ModuleName(modname string) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.modname = modname
		return o
	})
	return ob
}
