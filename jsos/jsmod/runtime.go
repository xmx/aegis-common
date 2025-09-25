package jsmod

import (
	"runtime"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewRuntime() jsvm.Module {
	return new(stdRuntime)
}

type stdRuntime struct{}

func (mod *stdRuntime) Preload(jsvm.Engineer) (string, any, bool) {
	vals := map[string]any{
		"memStats":     mod.memStats,
		"goos":         runtime.GOOS,
		"goarch":       runtime.GOARCH,
		"gc":           runtime.GC,
		"numCPU":       runtime.NumCPU,
		"numGoroutine": runtime.NumGoroutine,
		"numCgoCall":   runtime.NumCgoCall,
		"version":      runtime.Version,
	}

	return "runtime", vals, false
}

func (*stdRuntime) memStats() *runtime.MemStats {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	return stats
}
