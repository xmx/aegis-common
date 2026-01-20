package muxconn

import "sync/atomic"

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

type streamStat struct {
	history atomic.Int64
	active  atomic.Int64
}

func (ss *streamStat) openOne() {
	ss.history.Add(1)
	ss.active.Add(1)
}

func (ss *streamStat) closeOne() {
	ss.active.Add(-1)
}

func (ss *streamStat) NumStreams() (int64, int64) {
	return ss.history.Load(), ss.active.Load()
}
