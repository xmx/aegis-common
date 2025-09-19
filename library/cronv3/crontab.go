package cronv3

import (
	"context"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"
)

func New(parent context.Context, log *slog.Logger, opts ...cron.Option) *Crontab {
	return &Crontab{
		crond:  cron.New(opts...),
		uniq:   make(map[string]cron.EntryID, 16),
		log:    log,
		parent: parent,
	}
}

type Crontab struct {
	crond  *cron.Cron
	mutex  sync.Mutex
	uniq   map[string]cron.EntryID
	log    *slog.Logger
	parent context.Context
}

func (ctb *Crontab) Start() {
	ctb.crond.Start()
}

func (ctb *Crontab) Stop() {
	ctb.crond.Stop()
}

func (ctb *Crontab) AddTask(tsk Tasker) (bool, error) {
	info := tsk.Info()
	if info.ID == "" {
		info.ID = qualifiedID(tsk)
	}

	th := &taskHandler{
		info:   info,
		call:   tsk.Call,
		log:    ctb.log,
		parent: ctb.parent,
	}

	return ctb.addTask(th)
}

func (ctb *Crontab) AddJob(id, spec string, cmd func()) (bool, error) {
	if id == "" {
		id = qualifiedID(cmd)
	}

	th := &taskHandler{
		info: TaskInfo{
			ID:       id,
			CronSpec: spec,
		},
		call: func(context.Context) error {
			cmd()
			return nil
		},
		log:    ctb.log,
		parent: ctb.parent,
	}

	return ctb.addTask(th)
}

func (ctb *Crontab) AddSchedule(id string, sched cron.Schedule, cmd func()) (bool, error) {
	if id == "" {
		id = qualifiedID(cmd)
	}
	th := &taskHandler{
		info: TaskInfo{
			ID:        id,
			CronSched: sched,
		},
		call: func(context.Context) error {
			cmd()
			return nil
		},
		log:    ctb.log,
		parent: ctb.parent,
	}

	return ctb.addTask(th)
}

func (ctb *Crontab) Remove(id string) {
	ctb.mutex.Lock()
	_ = ctb.remove(id)
	ctb.mutex.Unlock()
}

func (ctb *Crontab) addTask(th *taskHandler) (bool, error) {
	info := th.info
	id, imm, sched := info.ID, info.Immediate, info.CronSched

	var err error
	var replace bool
	if sched != nil {
		replace = ctb.addSchedule(id, sched, th)
	} else {
		replace, err = ctb.addJob(id, info.CronSpec, th)
	}
	if err != nil || !imm {
		return replace, err
	}
	go th.Run()

	return replace, nil
}

// addJob 添加定时任务。
//
// - true:  名字已存在并替换原有的同名任务。
// - false: 名字不存在直接新增。
func (ctb *Crontab) addJob(id, spec string, job cron.Job) (bool, error) {
	ctb.mutex.Lock()
	eid, err := ctb.crond.AddJob(spec, job)
	if err != nil {
		ctb.mutex.Unlock()
		return false, err
	}

	lastID, exists := ctb.uniq[id]
	if exists {
		ctb.crond.Remove(lastID)
	}
	ctb.uniq[id] = eid
	ctb.mutex.Unlock()

	return exists, nil
}

// addSchedule 添加定时任务。
//
// - true:  名字已存在并替换原有的同名任务。
// - false: 名字不存在直接新增。
func (ctb *Crontab) addSchedule(id string, spec cron.Schedule, job cron.Job) bool {
	ctb.mutex.Lock()
	lastID, exists := ctb.uniq[id]
	if exists {
		ctb.crond.Remove(lastID)
	}
	newID := ctb.crond.Schedule(spec, job)
	ctb.uniq[id] = newID
	ctb.mutex.Unlock()

	return exists
}

// remove 通过名字删除定时任务。
func (ctb *Crontab) remove(name string) bool {
	if id, ok := ctb.uniq[name]; ok {
		ctb.crond.Remove(id)
		delete(ctb.uniq, name)
		return true
	}

	return false
}

// removeByID 通过 cron.EntryID 删除定时任务。
// 如果删除成功返回任务名和成功标志。
func (ctb *Crontab) removeByID(id cron.EntryID) (string, bool) {
	for name, eid := range ctb.uniq {
		if id == eid {
			ctb.remove(name)
			return name, true
		}
	}

	return "", false
}

// Cleanup 清理哪些不再执行的定时任务，该功能主要针对 [NewTimingSchedule] 类型的定时任务。
func (ctb *Crontab) cleanup() {
	ctb.mutex.Lock()
	for _, ent := range ctb.crond.Entries() {
		if ent.Next.IsZero() {
			_, _ = ctb.removeByID(ent.ID)
		}
	}
	ctb.mutex.Unlock()
}
