package jsmod_test

import (
	"context"
	"os"
	"testing"

	"github.com/xmx/aegis-common/jsmod"
	"github.com/xmx/jsos/jsvm"
)

func TestVariable(t *testing.T) {
	const filename = "testdata/variable.js"
	dat, err := os.ReadFile(filename)
	if err != nil {
		t.Error(err)
		return
	}
	code := string(dat)

	vm := newVM()
	varb := jsmod.NewVariable[VarConfig]("aegis/config")
	vm.Require().Register(varb)

	val, err := vm.RunScript(filename, code)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(val)
	t.Log(varb.Get())
}

func newVM() jsvm.Engineer {
	vm := jsvm.New(context.Background())
	vm.Require().Register(jsmod.Modules()...)
	stdout, _ := vm.Output()
	stdout.Attach(os.Stdout)

	return vm
}

type VarConfig struct {
	Addr string `json:"addr"`
}
