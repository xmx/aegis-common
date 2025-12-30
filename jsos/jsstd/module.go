package jsstd

import "github.com/xmx/aegis-common/jsos/jsvm"

func All() []jsvm.Module {
	return []jsvm.Module{
		NewConsole(),
		NewContext(),
		NewHTTP(),
		NewHTTPUtil(),
		NewNet(),
		NewOS(),
		NewRegexp(),
		NewRuntime(),
		NewTime(),
		NewURL(),
	}
}
