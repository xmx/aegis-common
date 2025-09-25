package jsvm

import (
	"github.com/evanw/esbuild/pkg/api"
	"github.com/grafana/sobek/file"
	"github.com/grafana/sobek/parser"
)

// Transform transpiles the input source string and strip types from it.
// this is done using esbuild
func Transform(name, code string) (string, error) {
	opts := api.TransformOptions{
		Loader:        api.LoaderJS,
		Sourcefile:    name,
		Target:        api.ESNext,
		Format:        api.FormatCommonJS,
		Sourcemap:     api.SourceMapInline,
		LegalComments: api.LegalCommentsNone,
		Platform:      api.PlatformNeutral,
		LogLevel:      api.LogLevelSilent,
		Charset:       api.CharsetUTF8,
	}

	ret := api.Transform(code, opts)
	if err := esbuildCheckError(ret); err != nil {
		return "", err
	}

	return string(ret.Code), nil
}

func esbuildCheckError(result api.TransformResult) error {
	es := make(parser.ErrorList, 0, len(result.Errors))
	for _, m := range result.Errors {
		var pos file.Position
		if l := m.Location; l != nil {
			pos.Filename = l.File
			pos.Line = l.Line
			pos.Column = l.Column
		}
		es.Add(pos, m.Text)
	}

	return es.Err()
}
