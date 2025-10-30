package profile

import (
	"encoding/json/v2"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Reader[T any] interface {
	// Read 读取配置文件。
	Read() (*T, error)
}

type File[T any] string

func (f File[T]) Read() (*T, error) {
	filename := string(f)
	ext := strings.ToLower(filepath.Ext(filename))
	if ext != ".json" && ext != ".jsonc" {
		return nil, errors.ErrUnsupported
	}

	fr, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fr.Close()

	t := new(T)
	if ext == ".json" {
		if err = json.UnmarshalRead(fr, t); err != nil {
			return nil, err
		}

		return t, nil
	}

	data, err := io.ReadAll(fr)
	if err != nil {
		return nil, err
	}
	bs := translate(data)
	if err = json.Unmarshal(bs, t); err != nil {
		return nil, err
	}

	return t, nil
}
