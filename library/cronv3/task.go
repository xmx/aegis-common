package cronv3

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
)

// Tasker 定时任务。
type Tasker interface {
	// Info 定时任务信息。
	Info() TaskInfo

	// Call 定时触发的实际函数。
	Call(ctx context.Context) error
}

type TaskInfo struct {
	ID        string        `json:"id"`                 // 任务唯一标识，如果为空则会通过反射获取 FQDN 作为唯一标识，此时同一个 struct 不同实例 FQDN 是一样的。
	Name      string        `json:"name,omitzero"`      // Name 任务名，简要描述任务说明。
	Immediate bool          `json:"immediate"`          // 添加到定时任务管理器时是否立即执行。
	Timeout   time.Duration `json:"timeout,omitzero"`   // 每次超时时间，小于等于零时代表无需设置。
	CronSched cron.Schedule `json:"-"`                  // 触发周期，等同于 CronSpec，但是优先级高于 CronSpec。
	CronSpec  string        `json:"cron_spec,omitzero"` // 触发周期，等同于 CronSched，但是当 CronSched 为空时才会生效。
}

func qualifiedID(obj any) string {
	tof := reflect.TypeOf(obj)
	kind := tof.Kind()
	if kind == reflect.Pointer {
		tof = tof.Elem()
	} else if kind == reflect.Func {
		vof := reflect.ValueOf(obj).Pointer()
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

type taskHandler struct {
	info    TaskInfo
	call    func(ctx context.Context) error
	log     *slog.Logger
	parent  context.Context
	running atomic.Bool
}

func (th *taskHandler) Run() {
	info := th.info
	attrs := []any{slog.Any("task_info", info)}
	if !th.running.CompareAndSwap(false, true) {
		th.log.Warn("定时任务已在运行中", attrs...)
		return
	}
	defer th.running.Store(false)

	ctx := th.parent
	timeout := info.Timeout
	if timeout >= 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	panicked, err := th.safeCall(ctx)
	if err == nil {
		th.log.Info("定时任务运行完毕", attrs...)
		return
	}
	attrs = append(attrs, slog.Any("error", err))
	if panicked {
		attrs = append(attrs, slog.Bool("panicked", true))
	}
	th.log.Error("定时任务运行出错", attrs...)
}

func (th *taskHandler) safeCall(ctx context.Context) (panicked bool, err error) {
	defer func() {
		if v := recover(); v != nil {
			panicked = true
			if e, ok := v.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", v)
			}
			debug.PrintStack()
		}
	}()

	err = th.call(ctx)

	return
}
