package tunutil

import (
	"encoding/binary"
	"encoding/json/v2"
	"io"
)

func ReadAuth(r io.Reader, v any) error {
	size := make([]byte, 2)
	if n, err := r.Read(size); err != nil {
		return err
	} else if n != 2 {
		return io.ErrShortBuffer
	}

	n := binary.BigEndian.Uint16(size)
	if n == 0 {
		return io.EOF
	}

	data := make([]byte, n)
	num, err := io.ReadFull(r, data)
	if err != nil {
		return err
	} else if num != int(n) {
		return io.ErrShortBuffer
	}

	return json.Unmarshal(data, v)
}

func WriteAuth(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	n := len(data)
	if n > 65535 {
		return io.ErrUnexpectedEOF
	}

	size := make([]byte, 2)
	binary.BigEndian.PutUint16(size, uint16(n))
	if _, err = w.Write(size); err == nil {
		_, err = w.Write(data)
	}

	return err
}
