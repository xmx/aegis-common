package cronv3

import (
	"slices"
	"time"

	"github.com/robfig/cron/v3"
)

// NewFixedTimes 在固定的几个时间点执行的定时任务。
func NewFixedTimes(times []time.Time) cron.Schedule {
	slices.SortFunc(times, func(a, b time.Time) int {
		return int(a.Sub(b))
	})

	return &fixedTimes{
		times: times,
	}
}

type fixedTimes struct {
	times []time.Time
}

func (ft *fixedTimes) Next(now time.Time) time.Time {
	for idx, at := range ft.times {
		if at.After(now) {
			ft.times = ft.times[idx:]
			return at
		}
	}

	return time.Time{}
}
