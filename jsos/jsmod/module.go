package jsmod

import "github.com/xmx/aegis-common/jsos/jsvm"

func Modules() []jsvm.Module {
	return []jsvm.Module{
		NewConsole(),
		NewContext(),
		NewNet(),
		NewOS(),
		NewRuntime(),
		NewTime(),
		NewURL(),
	}
}
