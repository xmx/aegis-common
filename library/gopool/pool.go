package gopool

import (
	"context"
	"sync/atomic"
)

type Pool interface {
	Go(fn func()) (done <-chan struct{})
	Gos(parent context.Context, fns ...func(parent context.Context)) (done <-chan struct{})
}

func New(worker int) Pool {
	if worker <= 0 {
		worker = 1
	}
	return &pool{
		sema: make(chan struct{}, worker),
	}
}

type pool struct {
	sema chan struct{}
}

func (p *pool) Go(fn func()) <-chan struct{} {
	done := make(chan struct{})
	if fn == nil {
		close(done)
		return done
	}

	fun := p.warp(done, fn)
	p.sched(fun)

	return done
}

func (p *pool) Gos(parent context.Context, fns ...func(parent context.Context)) <-chan struct{} {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)
	mon := p.newMonitor(ctx, cancel)
	funcs := make([]func(), 0, len(fns))
	for _, f := range fns {
		if f != nil {
			fn := mon.warp(f)
			funcs = append(funcs, fn)
		}
	}
	if len(funcs) == 0 {
		cancel()
		return ctx.Done()
	}
	for _, fn := range funcs {
		p.sched(fn)
	}

	return ctx.Done()
}

func (p *pool) sched(fn func()) {
	p.sema <- struct{}{}
	go fn()
}

func (p *pool) warp(done chan struct{}, fn func()) func() {
	return func() {
		defer func() {
			close(done)
			<-p.sema
		}()
		fn()
	}
}

func (p *pool) newMonitor(ctx context.Context, cancel context.CancelFunc) *monitor {
	return &monitor{sema: p.sema, ctx: ctx, cancel: cancel}
}

type monitor struct {
	sema   chan struct{}
	cnt    atomic.Int32
	ctx    context.Context
	cancel context.CancelFunc
}

func (m *monitor) warp(fn func(context.Context)) func() {
	m.cnt.Add(1)
	return func() {
		defer func() {
			if num := m.cnt.Add(-1); num <= 0 {
				m.cancel()
			}
			<-m.sema
		}()
		fn(m.ctx)
	}
}
