package jsexec

import (
	"io"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

type option struct {
	stdout  io.Writer
	stderr  io.Writer
	modules []jsvm.Module
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

func (ob OptionBuilder) Stdout(w io.Writer) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.stdout = w
		return o
	})
	return ob
}

func (ob OptionBuilder) Stderr(w io.Writer) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.stderr = w
		return o
	})
	return ob
}

func (ob OptionBuilder) Modules(mods []jsvm.Module) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.modules = mods
		return o
	})
	return ob
}
