package jsmod_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/robfig/cron/v3"
	"github.com/xmx/aegis-common/jsos/jsmod"
	"github.com/xmx/aegis-common/jsos/jsvm"
	"github.com/xmx/aegis-common/library/cronv3"
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
	crond := cronv3.New(context.Background(), slog.Default(), cron.WithSeconds())
	crond.Start()
	crontab := jsmod.NewCrontab(crond)
	vm.Require().Register(crontab)

	val, err := vm.RunScript(filename, code)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log(val)
}

func newVM() jsvm.Engineer {
	vm := jsvm.New(context.Background())
	vm.Require().Registers(jsmod.Modules())
	stdout, _ := vm.Output()
	stdout.Attach(os.Stdout)

	return vm
}

func TestPrint(t *testing.T) {
	vm := newVM()
	vm.Require().Registers(jsmod.Modules())
	stdout, _ := vm.Output()
	stdout.Attach(os.Stdout)

	vm.RunScript("s", "import console from 'console'\nconsole.log('hello world')\nconsole.log('hello world')")
}
