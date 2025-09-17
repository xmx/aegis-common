package transport

import "sync/atomic"

// MuxLoader 对于客户端，存在掉线重连问题，那么重连成功后
// 要保证应用上层在持有对象/指针不改变的情况下，安全的替换新的连接。
type MuxLoader interface {
	// LoadMux 获取多路复用连接。
	LoadMux() (Muxer, bool)

	// StoreMux 替换/存放一个多路复用连接。
	StoreMux(Muxer)
}

func NewMuxLoader(m Muxer) MuxLoader {
	l := new(muxerLoader)
	l.StoreMux(m)

	return l
}

type muxerLoader struct {
	v atomic.Pointer[muxerHolder]
}

func (l *muxerLoader) StoreMux(m Muxer) {
	if m != nil {
		h := &muxerHolder{m: m}
		l.v.Store(h)
	}
}

func (l *muxerLoader) LoadMux() (Muxer, bool) {
	if h := l.v.Load(); h != nil {
		if m := h.m; m != nil {
			return m, true
		}
	}

	return nil, false
}

type muxerHolder struct {
	m Muxer
}
