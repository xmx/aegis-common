package cronv3

import (
	"context"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"
)

type Crontab struct {
	log   *slog.Logger
	ctab  *cron.Cron
	mutex sync.RWMutex
	tasks map[string]*scheduledTask
}

func New(log *slog.Logger, opts ...cron.Option) *Crontab {
	return &Crontab{
		log:   log,
		ctab:  cron.New(opts...),
		tasks: make(map[string]*scheduledTask, 8),
	}
}

func (c *Crontab) Start() {
	c.ctab.Start()
}

func (c *Crontab) Stop() {
	c.ctab.Stop()
}

func (c *Crontab) Clean() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for id, task := range c.tasks {
		entryID := task.entryID
		entry := c.ctab.Entry(entryID)
		if entry.Next.IsZero() {
			c.remove(id)
		}
	}
}

func (c *Crontab) AddTask(task Tasker) error {
	info := task.Info()
	if info.ID == "" {
		info.ID = qualifiedID(task)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.addFunc(info, task.Call)
}

func (c *Crontab) AddJob(spec string, cmd func()) error {
	id := qualifiedID(cmd)
	info := TaskInfo{ID: id, CronSpec: spec}
	fn := wrapFunc(cmd)

	return c.addFunc(info, fn)
}

func (c *Crontab) AddSchedule(sch cron.Schedule, cmd func()) error {
	id := qualifiedID(cmd)
	info := TaskInfo{ID: id, CronSched: sch}
	fn := wrapFunc(cmd)

	return c.addFunc(info, fn)
}

func (c *Crontab) Remove(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.remove(id)
}

func (c *Crontab) Tasks() []Tasker {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var ret []Tasker
	for _, task := range c.tasks {
		ret = append(ret, task)
	}

	return ret
}

func (c *Crontab) remove(id string) {
	if last := c.tasks[id]; last != nil {
		c.ctab.Remove(last.entryID)
		delete(c.tasks, id)
	}
}

func (c *Crontab) addFunc(info TaskInfo, exec func(context.Context) error) error {
	if info.ID == "" {
		info.ID = qualifiedID(exec)
	}

	id := info.ID
	sch := &scheduledTask{info: info, exec: exec, log: c.log}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.remove(id) // 删除原来的任务（如果存在）

	if info.CronSched != nil {
		sch.entryID = c.ctab.Schedule(info.CronSched, sch)
	} else {
		entryID, err := c.ctab.AddJob(info.CronSpec, sch)
		if err != nil {
			return err
		}
		sch.entryID = entryID
	}
	c.tasks[id] = sch

	if info.Immediate { // 立即执行
		go sch.Run()
	}

	return nil
}
