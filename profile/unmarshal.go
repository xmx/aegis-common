package profile

import (
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/xmx/aegis-common/options"
)

type Reader[T any] interface {
	Read(ctx context.Context) (*T, error)
}

func NewFile[T any](filename string, opts ...options.Lister[option]) Reader[T] {
	opt := options.Eval(opts...)

	return &fileUnmarshal[T]{
		filename: filename,
		opt:      opt,
	}
}

type fileUnmarshal[T any] struct {
	filename string
	opt      option
}

func (f *fileUnmarshal[T]) Read(ctx context.Context) (*T, error) {
	fn := f.filename
	stat, err := os.Stat(fn)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return f.unmarshalDir(ctx, fn)
	}

	return f.unmarshalFile(ctx, fn)
}

func (f *fileUnmarshal[T]) unmarshalFile(ctx context.Context, filename string) (*T, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	r := f.limitReader(fd)

	t := new(T)
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".js", ".javascript":
		err = executeJS(ctx, r, t, f.opt)
	case ".json":
		err = json.UnmarshalRead(r, t)
	case ".jsonc":
		err = unmarshalJSONC(r, t)
	default:
		err = errors.ErrUnsupported
	}
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (f *fileUnmarshal[T]) unmarshalDir(ctx context.Context, basedir string) (*T, error) {
	// 按照 .js .jsonc .json 顺序读取配置文件，直到第一个正确的停止。
	errs := make([]error, 0, 3)
	basename := "application"
	names := []string{basename + ".js", basename + ".jsonc", basename + ".json"}
	for _, name := range names {
		cfg, err := f.unmarshalFile(ctx, filepath.Join(basedir, name))
		if err == nil {
			return cfg, nil
		}
		errs = append(errs, err)
	}

	return nil, errors.Join(errs...)
}

func (f *fileUnmarshal[T]) limitReader(r io.Reader) io.Reader {
	if n := f.opt.limit; n > 0 {
		return io.LimitReader(r, n)
	}

	return r
}
