package jsmod_test

import (
	"context"
	"os"
	"testing"

	"github.com/xmx/aegis-common/jsos/jsmod"
	"github.com/xmx/aegis-common/jsos/jsvm"
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
	val, err := vm.RunScript(filename, code)
	if err != nil {
		t.Log(err)
		return
	}

	t.Log(val)
}

func newVM() jsvm.Engineer {
	vm := jsvm.NewVM(context.Background())
	vm.Require().Registers(jsmod.All()...)
	stdout, _ := vm.Output()
	stdout.Append(os.Stdout)

	return vm
}
