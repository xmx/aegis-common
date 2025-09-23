package wsocket

import (
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func NewUpgrade() *websocket.Upgrader {
	return &websocket.Upgrader{
		HandshakeTimeout:  10 * time.Second,
		CheckOrigin:       func(*http.Request) bool { return true },
		EnableCompression: true,
	}
}

type PipeResult struct {
	AtoBCount int64 `json:"atob_count"`
	BtoACount int64 `json:"btoa_count"`
	AtoBError error `json:"atob_error"`
	BtoAError error `json:"btoa_error"`
}

func Pipe(a, b *websocket.Conn) PipeResult {
	atob := goPipeCopy(b, a)
	btoa := pipeCopy(a, b)

	return PipeResult{
		AtoBCount: atob.cnt,
		BtoACount: btoa.cnt,
		AtoBError: atob.err,
		BtoAError: btoa.err,
	}
}

type pipeResult struct {
	cnt int64
	err error
}

func pipeCopy(dst, src *websocket.Conn) *pipeResult {
	ret := new(pipeResult)
	for {
		mt, rd, err := src.NextReader()
		if err != nil {
			ret.err = err
			_ = dst.Close()
			break
		}
		wt, err1 := dst.NextWriter(mt)
		if err1 != nil {
			ret.err = err1
			_ = src.Close()
			break
		}

		n, err2 := ioCopy(wt, rd)
		ret.cnt += n
		if err2 != nil {
			ret.err = err2
			break
		}
	}

	return ret
}

func ioCopy(dst io.WriteCloser, src io.Reader) (int64, error) {
	//goland:noinspection GoUnhandledErrorResult
	defer dst.Close()
	return io.Copy(dst, src)
}

func goPipeCopy(dst, src *websocket.Conn) *pipeResult {
	ch := make(chan *pipeResult, 1)
	go func() { ch <- pipeCopy(dst, src) }()
	ret := <-ch

	return ret
}
