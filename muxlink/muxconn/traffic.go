package muxconn

import (
	"context"
	"errors"
	"io"
	"math"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

type trafficStat struct {
	rx, tx atomic.Uint64
}

func (ts *trafficStat) Load() (rx, tx uint64) {
	return ts.rx.Load(), ts.tx.Load()
}

func (ts *trafficStat) incrRX(n int) {
	if n > 0 {
		ts.rx.Add(uint64(n))
	}
}

func (ts *trafficStat) incrTX(n int) {
	if n > 0 {
		ts.tx.Add(uint64(n))
	}
}

const (
	minimumBurst     = 2 << 14
	defaultLimitWait = 10 * time.Second
)

func newUnlimit() *rateLimiter {
	return &rateLimiter{
		rlimit: rate.NewLimiter(rate.Inf, minimumBurst),
		wlimit: rate.NewLimiter(rate.Inf, minimumBurst),
	}
}

type rateLimiter struct {
	rlimit *rate.Limiter
	wlimit *rate.Limiter
}

func (rl *rateLimiter) Limit() rate.Limit { return rl.rlimit.Limit() }

func (rl *rateLimiter) SetLimit(bps rate.Limit) {
	if bps < minimumBurst {
		bps = minimumBurst
	}

	now := time.Now()
	rl.rlimit.SetLimitAt(now, bps)
	rl.wlimit.SetLimitAt(now, bps)

	if bps != rate.Inf {
		burst := int(math.Ceil(float64(bps) * 1.2))
		burst = max(burst, minimumBurst)
		rl.rlimit.SetBurstAt(now, burst)
		rl.wlimit.SetBurstAt(now, burst)
	}
}

func (rl *rateLimiter) newReadWriter(parent context.Context, rw io.ReadWriter) io.ReadWriter {
	if parent == nil {
		parent = context.Background()
	}

	return &limitReadWriter{
		rlimit: rl.rlimit,
		wlimit: rl.wlimit,
		under:  rw,
		parent: parent,
	}
}

type limitReadWriter struct {
	rlimit *rate.Limiter
	wlimit *rate.Limiter
	under  io.ReadWriter
	parent context.Context
}

func (lrw *limitReadWriter) Read(p []byte) (int, error) {
	if len(p) == 0 || lrw.rlimit.Limit() == rate.Inf {
		return lrw.under.Read(p)
	}

	tokens := min(len(p), minimumBurst)
	if err := lrw.waitN(lrw.rlimit, tokens); err != nil {
		return 0, err
	}

	return lrw.under.Read(p[:tokens])
}

func (lrw *limitReadWriter) Write(p []byte) (int, error) {
	total := len(p)
	if total == 0 || lrw.wlimit.Limit() == rate.Inf {
		return lrw.under.Write(p)
	}

	remain := total
	var written int
	for remain > 0 {
		tokens := min(remain, minimumBurst)
		if err := lrw.waitN(lrw.wlimit, tokens); err != nil {
			return written, err
		}

		n, err := lrw.under.Write(p[written : written+tokens])
		written += n
		if err != nil {
			return written, err
		}

		remain = total - written
	}

	return written, nil
}

func (lrw *limitReadWriter) waitN(limit *rate.Limiter, n int) error {
	ctx, cancel := context.WithTimeout(lrw.parent, defaultLimitWait)
	defer cancel()

	if err := limit.WaitN(ctx, n); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return os.ErrDeadlineExceeded
		}

		return err
	}

	return nil
}
