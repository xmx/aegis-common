package profile

import (
	"bytes"
	"encoding/json"
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

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err
	}

	return os.WriteFile(name, buf.Bytes(), 0o600)
}
