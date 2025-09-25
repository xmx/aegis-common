package jsmod

import (
	"os"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

func NewOS() jsvm.Module {
	return new(stdOS)
}

type stdOS struct{}

func (mod *stdOS) Preload(eng jsvm.Engineer) (string, any, bool) {
	vals := map[string]any{
		"getpid":       os.Getpid,
		"open":         os.Open,
		"hostname":     os.Hostname,
		"tempDir":      os.TempDir,
		"getenv":       os.Getenv,
		"userCacheDir": os.UserCacheDir,
		"environ":      os.Environ,
		"expand":       os.Expand,
		"expandEnv":    os.ExpandEnv,
		"getuid":       os.Getuid,
		"geteuid":      os.Geteuid,
		"getgid":       os.Getgid,
		"getegid":      os.Getegid,
		"getgroups":    os.Getgroups,
		"getpagesize":  os.Getpagesize,
		"getppid":      os.Getppid,
		"getwd":        os.Getwd,
	}

	return "os", vals, false
}
