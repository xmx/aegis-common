package transport

import "context"

type Peer[K comparable] interface {
	ID() K
	Mux() Muxer
}

func WithValue[K comparable](ctx context.Context, p Peer[K]) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if p == nil {
		return ctx
	}

	return context.WithValue(ctx, defaultContextKey, p)
}

func FromContext[K comparable](ctx context.Context) Peer[K] {
	if ctx == nil {
		return nil
	}

	if p, ok := ctx.Value(defaultContextKey).(Peer[K]); ok {
		return p
	}

	return nil
}

var defaultContextKey = contextKey{}

type contextKey struct{}

func (contextKey) String() string {
	return "transport-peer-context-key"
}
