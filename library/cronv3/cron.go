package cronv3

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

func New(opts ...cron.Option) *Crontab {
	return &Crontab{
		crond: cron.New(opts...),
		names: make(map[string]cron.EntryID, 16),
	}
}

type Crontab struct {
	crond *cron.Cron
	mutex sync.Mutex
	names map[string]cron.EntryID
}

func (c *Crontab) Start() {
	c.crond.Start()
}

func (c *Crontab) Run() {
	c.crond.Run()
}

func (c *Crontab) Stop() context.Context {
	return c.crond.Stop()
}

func (c *Crontab) Location() *time.Location {
	return c.crond.Location()
}

// AddJob 添加定时任务。
//
// - true:  名字已存在并替换原有的同名任务。
// - false: 名字不存在直接新增。
func (c *Crontab) AddJob(name, spec string, cmd func()) (bool, error) {
	job := c.newJob(name, cmd)

	c.mutex.Lock()
	id, err := c.crond.AddJob(spec, job)
	if err != nil {
		c.mutex.Unlock()
		return false, err
	}

	lastID, exists := c.names[name]
	if exists {
		c.crond.Remove(lastID)
	}
	c.names[name] = id
	c.mutex.Unlock()

	return exists, nil
}

// AddSchedule 添加定时任务。
//
// - true:  名字已存在并替换原有的同名任务。
// - false: 名字不存在直接新增。
func (c *Crontab) AddSchedule(name string, spec cron.Schedule, cmd func()) bool {
	job := c.newJob(name, cmd)

	c.mutex.Lock()
	lastID, exists := c.names[name]
	if exists {
		c.crond.Remove(lastID)
	}
	newID := c.crond.Schedule(spec, job)
	c.names[name] = newID
	c.mutex.Unlock()

	return exists
}

func (c *Crontab) Remove(name string) {
	c.mutex.Lock()
	_ = c.remove(name)
	c.mutex.Unlock()
}

// Cleanup 清理哪些不再执行的定时任务，该功能主要针对 [NewTimingSchedule] 类型的定时任务。
func (c *Crontab) cleanup() {
	c.mutex.Lock()
	for _, ent := range c.crond.Entries() {
		if ent.Next.IsZero() {
			_, _ = c.removeByID(ent.ID)
		}
	}
	c.mutex.Unlock()
}

// remove 通过名字删除定时任务。
func (c *Crontab) remove(name string) bool {
	if id, ok := c.names[name]; ok {
		c.crond.Remove(id)
		delete(c.names, name)
		return true
	}

	return false
}

// removeByID 通过 cron.EntryID 删除定时任务。
// 如果删除成功返回任务名和成功标志。
func (c *Crontab) removeByID(id cron.EntryID) (string, bool) {
	for name, eid := range c.names {
		if id == eid {
			c.remove(name)
			return name, true
		}
	}

	return "", false
}

func (c *Crontab) newJob(name string, cmd func()) cron.Job {
	return &cronJob{
		name: name,
		cmd:  cmd,
	}
}

type cronJob struct {
	name string
	cmd  func()
}

func (cj *cronJob) Run() {
	defer func() {
		if v := recover(); v != nil {
			_, _ = fmt.Fprintf(os.Stderr, "job name: %s, %v\n", cj.name, v)
			debug.PrintStack()
		}
	}()

	cj.cmd()
}
