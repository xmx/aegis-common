package jstask

import (
	"io"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

type Option struct {
	Stdout  io.Writer
	Stderr  io.Writer
	Modules []jsvm.Module
}
