package transport

import (
	"encoding/json"
	"io"
)

type AuthRequest struct {
	ID     string `json:"id"     validate:"required"`
	Goos   string `json:"goos"   validate:"required,oneof=darwin dragonfly illumos ios js wasip1 linux android solaris freebsd nacl netbsd openbsd plan9 windows aix"`
	Goarch string `json:"goarch" validate:"required,oneof=386 amd64 arm arm64 loong64 mips mipsle mips64 mips64le ppc64 ppc64le riscv64 s390x sparc64 wasm"`
	Secret string `json:"secret" validate:"required,lte=500"`
}

func (v *AuthRequest) ReadFrom(r io.Reader) (int64, error) {
	wc := new(writeN)
	rd := io.TeeReader(r, wc)
	err := json.NewDecoder(rd).Decode(v)

	return wc.N(), err
}

func (v *AuthRequest) WriteTo(w io.Writer) (int64, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	n, e := w.Write(b)

	return int64(n), e
}

type AuthResponse struct {
	Succeed bool   `json:"succeed,omitzero"`
	Message string `json:"message,omitzero"`
}

func (v *AuthResponse) ReadFrom(r io.Reader) (int64, error) {
	wc := new(writeN)
	rd := io.TeeReader(r, wc)
	err := json.NewDecoder(rd).Decode(v)

	return wc.N(), err
}

func (v *AuthResponse) WriteTo(w io.Writer) (int64, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return 0, err
	}
	n, e := w.Write(b)

	return int64(n), e
}

func (v *AuthResponse) Error() string {
	if v.Succeed {
		return "<nil>"
	}

	return v.Message
}

type writeN struct {
	n int64
}

func (w *writeN) Write(p []byte) (int, error) {
	n := len(p)
	w.n += int64(n)
	return n, nil
}

func (w *writeN) N() int64 {
	return w.n
}
