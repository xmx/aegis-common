package cronv3

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

type TaskInfo struct {
	ID        string        `json:"id"`                  // 任务唯一标识，如果为空则会通过反射获取 FQDN 作为唯一标识，此时同一个 struct 不同实例 FQDN 是一样的。
	Name      string        `json:"name"`                // Name 任务名，简要描述任务说明。
	Immediate bool          `json:"immediate,omitzero"`  // 添加到定时任务管理器时是否立即执行。
	Timeout   time.Duration `json:"timeout"`             // 每次超时时间，小于等于零时代表无需设置。
	CronSched cron.Schedule `json:"cron_sched,omitzero"` // 触发周期，等同于 CronSpec，但是优先级高于 CronSpec。
	CronSpec  string        `json:"cron_spec,omitzero"`  // 触发周期，等同于 CronSched，但是当 CronSched 为空时才会生效。
}

// Tasker 定时任务。
type Tasker interface {
	// Info 定时任务信息。
	Info() TaskInfo

	// Call 定时触发的实际函数。
	Call(ctx context.Context) error
}

type scheduledTask struct {
	log     *slog.Logger
	entryID cron.EntryID
	working atomic.Bool // 当前任务是否
	info    TaskInfo    // 任务信息
	exec    func(context.Context) error
}

func (s *scheduledTask) Run() {
	_ = s.Call(context.Background())
}

func (s *scheduledTask) Info() TaskInfo {
	return s.info
}

func (s *scheduledTask) Call(ctx context.Context) error {
	if !s.working.CompareAndSwap(false, true) {
		return nil
	}
	defer s.working.Store(false)

	startedAt := time.Now()
	panicked, err := s.safeCall(ctx)
	attrs := []any{"info", s.info, "started_at", startedAt, "finished_at", time.Now()}
	if err == nil {
		s.log.Debug("定时任务执行完毕", attrs...)
		return nil
	}
	if panicked {
		s.log.Error("定时任务执行发生 panic", attrs...)
	} else {
		s.log.Warn("定时任务执行出错", attrs...)
	}

	return err
}

func (s *scheduledTask) safeCall(ctx context.Context) (panicked bool, err error) {
	if du := s.info.Timeout; du > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, du)
		defer cancel()
	}

	defer func() {
		if v := recover(); v != nil {
			panicked = true
			if ve, ok := v.(error); ok {
				err = ve
			} else {
				err = fmt.Errorf("cron task panicked: %v", v)
			}
			debug.PrintStack()
		}
	}()

	err = s.exec(ctx)

	return
}

func wrapFunc(cmd func()) func(context.Context) error {
	return func(context.Context) error {
		cmd()
		return nil
	}
}

func qualifiedID(v any) string {
	s := fqdn(v)
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func fqdn(v any) string {
	tof := reflect.TypeOf(v)
	kind := tof.Kind()
	if kind == reflect.Pointer {
		tof = tof.Elem()
	} else if kind == reflect.Func {
		vof := reflect.ValueOf(v).Pointer()
		if pc := runtime.FuncForPC(vof); pc != nil {
			return pc.Name()
		}

		return ""
	}

	pkg, name := tof.PkgPath(), tof.Name()
	if pkg == "" {
		pkg = "main"
	}

	return pkg + "." + name
}
