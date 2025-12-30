package jsstd

import (
	"regexp"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewRegexp() jsvm.Module {
	return new(regexpModule)
}

type regexpModule struct {
	svm jsvm.Engineer
}

func (r *regexpModule) Preload(svm jsvm.Engineer) (string, any, bool) {
	r.svm = svm
	vals := map[string]any{
		"compile":      regexp.Compile,
		"compilePOSIX": regexp.CompilePOSIX,
		"matchString":  regexp.MatchString,
		"quoteMeta":    regexp.QuoteMeta,
	}

	return "regexp", vals, true
}
