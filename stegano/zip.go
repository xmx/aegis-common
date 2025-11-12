package stegano

import (
	"archive/zip"
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"io"
	"io/fs"
)

const manifestName = "manifest.json"

func AddFS(w io.Writer, fsys fs.FS, offset int64) error {
	zw := zip.NewWriter(w)
	defer zw.Close()
	if offset > 0 {
		zw.SetOffset(offset)
	}

	return zw.AddFS(fsys)
}

func CreateManifestZip(manifest any, offset int64) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	defer zw.Close()
	if offset > 0 {
		zw.SetOffset(offset)
	}
	cw, err := zw.Create(manifestName)
	if err != nil {
		return nil, err
	}
	opt := jsontext.WithIndent("  ")
	if err = json.MarshalWrite(cw, manifest, opt); err != nil {
		return nil, err
	}
	if err = zw.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}

func Open(f string) (*zip.ReadCloser, error) {
	return zip.OpenReader(f)
}

func ReadManifest(f string, v any) error {
	zrc, err := Open(f)
	if err != nil {
		return err
	}
	defer zrc.Close()

	mf, err := zrc.Open(manifestName)
	if err != nil {
		return err
	}
	defer mf.Close()

	return json.UnmarshalRead(mf, v)
}

type File[T any] string

func (f File[T]) Read() (*T, error) {
	t := new(T)
	if err := ReadManifest(string(f), t); err != nil {
		return nil, err
	}

	return t, nil
}
