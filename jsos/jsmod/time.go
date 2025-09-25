package jsmod

import (
	"time"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewTime(debug ...bool) jsvm.Module {
	mod := new(stdTime)
	if len(debug) > 0 {
		mod.dbg = debug[0]
	}

	return mod
}

type stdTime struct {
	dbg bool
	eng jsvm.Engineer
}

func (mod *stdTime) Preload(eng jsvm.Engineer) (string, any, bool) {
	mod.eng = eng

	vals := map[string]any{
		"nanosecond ":   time.Nanosecond,
		"microsecond":   time.Microsecond,
		"millisecond":   time.Millisecond,
		"second":        time.Second,
		"minute":        time.Minute,
		"hour":          time.Hour,
		"january":       time.January,
		"february":      time.February,
		"march":         time.March,
		"april":         time.April,
		"may":           time.May,
		"june":          time.June,
		"july":          time.July,
		"august":        time.August,
		"september":     time.September,
		"october":       time.October,
		"november":      time.November,
		"december":      time.December,
		"sunday":        time.Sunday,
		"monday":        time.Monday,
		"tuesday":       time.Tuesday,
		"wednesday":     time.Wednesday,
		"thursday":      time.Thursday,
		"friday":        time.Friday,
		"saturday":      time.Saturday,
		"sleep":         mod.sleep,
		"local":         time.Local,
		"parseDuration": time.ParseDuration,
		"afterFunc":     time.AfterFunc,
	}

	return "time", vals, true
}

func (mod *stdTime) sleep(d time.Duration) {
	if !mod.dbg {
		return
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	ctx := mod.eng.Context()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}
