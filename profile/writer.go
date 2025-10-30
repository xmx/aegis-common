package profile

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"
)

func WriteFile(name string, v any) error {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.MarshalWrite(file, v, jsontext.WithIndent("  "))
}
