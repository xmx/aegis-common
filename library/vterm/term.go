package vterm

import "io"

type Typewriter interface {
	io.ReadWriteCloser
	Size() (Winsize, error)
	Setsize(Winsize) error
}

type Winsize struct {
	Cols uint32
	Rows uint32
}

func (w Winsize) IsZero() bool {
	return w.Cols == 0 && w.Rows == 0
}
