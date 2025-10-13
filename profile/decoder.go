package profile

import (
	"context"
	"encoding/json/v2"
	"io"
	"os"

	"github.com/xmx/aegis-common/jsos/jsmod"
	"github.com/xmx/aegis-common/jsos/jsvm"
)

func unmarshalJSONC[T any](r io.Reader, dst *T) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dst)
}

func executeJS[T any](ctx context.Context, r io.Reader, dst *T, opt option) error {
	code, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	modname := opt.modname
	if modname == "" {
		modname = "config"
	}
	data := &jsData[T]{
		modname: modname,
		data:    dst,
	}

	vm := jsvm.New(ctx)
	stdout, stderr := vm.Output()
	stdout.Attach(os.Stdout)
	stderr.Attach(os.Stderr)
	require := vm.Require()
	require.Registers(jsmod.Modules())
	require.Registers(opt.modules)
	require.Register(data)
	_, err = vm.RunScript(modname, string(code))

	return err
}

type jsData[T any] struct {
	modname string
	data    *T
}

func (j *jsData[T]) Preload(jsvm.Engineer) (string, any, bool) {
	vals := map[string]any{
		"get": j.get,
	}

	return j.modname, vals, true
}

func (j *jsData[T]) get() *T {
	return j.data
}
