package jsmod

import (
	"net/url"

	"github.com/xmx/jsos/jsvm"
)

func NewURL() jsvm.Module {
	return new(stdURL)
}

type stdURL struct{}

func (*stdURL) Preload(jsvm.Engineer) (string, any, bool) {
	vals := map[string]any{
		"parse":           url.Parse,
		"joinPath":        url.JoinPath,
		"parseQuery":      url.ParseQuery,
		"parseRequestURI": url.ParseRequestURI,
		"pathEscape":      url.PathEscape,
		"pathUnescape":    url.PathUnescape,
		"queryEscape":     url.QueryEscape,
		"queryUnescape":   url.QueryUnescape,
		"userPassword":    url.UserPassword,
	}

	return "net/url", vals, false
}
