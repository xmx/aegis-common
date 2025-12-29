package fsnotify

import (
	"github.com/xmx/aegis-common/jsos/jsvm"
)

func New() jsvm.Module {
	return new(fsnotifyPackage)
}

type fsnotifyPackage struct {
	svm jsvm.Engineer
}

func (f *fsnotifyPackage) Preload(svm jsvm.Engineer) (string, any, bool) {
	f.svm = svm
	return "", nil, false
}
