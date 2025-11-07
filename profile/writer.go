package profile

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"
	"path/filepath"
)

func WriteFile(name string, v any) error {
	dir := filepath.Dir(name)
	if _, err := os.Stat(dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err = os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	opts := []json.Options{
		jsontext.EscapeForHTML(false),
		jsontext.WithIndent("  "),
	}
	out, err := json.Marshal(v, opts...)
	if err != nil {
		return err
	}

	return os.WriteFile(name, out, 0o600)
}
