package transport

type Peer[K comparable] interface {
	ID() K
	Mux() Muxer
}
