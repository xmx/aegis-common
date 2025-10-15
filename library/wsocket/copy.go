package wsocket

import (
	"io"

	"github.com/gorilla/websocket"
)

type CopyStat struct {
	AtoBCount int64 `json:"atob_count"`
	BtoACount int64 `json:"btoa_count"`
	AtoBError error `json:"atob_error"`
	BtoAError error `json:"btoa_error"`
}

func ProxyCopy(a, b *websocket.Conn) CopyStat {
	var stat CopyStat
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		stat.BtoACount, stat.BtoAError = copyTo(a, b)
	}()

	stat.AtoBCount, stat.AtoBError = copyTo(b, a)

	<-wait

	return stat
}

func copyTo(dst, src *websocket.Conn) (int64, error) {
	var num int64
	for {
		mt, r, err := src.NextReader()
		if err != nil {
			_ = dst.Close()
			return num, err
		}

		w, err1 := dst.NextWriter(mt)
		if err1 != nil {
			_ = src.Close()
			return num, err1
		}

		n, err3 := io.Copy(w, r)
		_ = w.Close()
		if err3 != nil {
			_ = dst.Close()
			return num, err3
		}

		num += n
	}
}
