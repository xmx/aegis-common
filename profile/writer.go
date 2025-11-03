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

	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.MarshalWrite(file, v, jsontext.WithIndent("  "))
}
