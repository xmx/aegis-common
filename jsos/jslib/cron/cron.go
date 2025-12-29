package cron

import (
	"sync"
	"time"

	"github.com/grafana/sobek"
	"github.com/robfig/cron/v3"
	"github.com/xmx/aegis-common/jsos/jsvm"
)

func New() jsvm.Module {
	return new(cronPackage)
}

type cronPackage struct {
	svm jsvm.Engineer
	mtx sync.Mutex
}

func (cp *cronPackage) Preload(svm jsvm.Engineer) (string, any, bool) {
	cp.svm = svm
	vals := map[string]any{
		"secondOptional": cron.SecondOptional,
		"second":         cron.Second,
		"minute":         cron.Minute,
		"hour":           cron.Hour,
		"dom":            cron.Dom,
		"month":          cron.Month,
		"dow":            cron.Dow,
		"descriptor":     cron.Descriptor,
		"withSeconds":    cron.WithSeconds,
		"withLocation":   cron.WithLocation,
		"withParser":     cron.WithParser,
		"newParser":      cron.NewParser,
		"every":          cron.Every,
		"parseStandard":  cron.ParseStandard,
		"new":            cp.newCron,
	}

	return "github.com/robfig/cron/v3", vals, true
}

func (cp *cronPackage) serializer(job cron.Job) cron.Job {
	return cron.FuncJob(func() {
		cp.mtx.Lock()
		defer cp.mtx.Unlock()

		job.Run()
	})
}

func (cp *cronPackage) newCron(opts ...cron.Option) (*sobek.Object, error) {
	opts = append(opts, cron.WithChain(cp.serializer)) // 串行执行。
	ctab := cron.New(opts...)
	ct := &cronTab{ctab: ctab}
	cp.svm.Defer().Append(ct.close)

	obj := cp.svm.Runtime().NewObject()
	if err := obj.Set("run", ct.run); err != nil {
		return nil, err
	}
	if err := obj.Set("addJob", ct.addJob); err != nil {
		return nil, err
	}
	if err := obj.Set("schedule", ct.schedule); err != nil {
		return nil, err
	}
	if err := obj.Set("remove", ct.remove); err != nil {
		return nil, err
	}
	if err := obj.Set("entries", ct.entries); err != nil {
		return nil, err
	}

	return nil, nil
}

type cronTab struct {
	ctab *cron.Cron
}

func (ct *cronTab) run() {
	ct.ctab.Run()
}

func (ct *cronTab) close() error {
	ct.ctab.Stop()
	return nil
}

func (ct *cronTab) addJob(spec string, cmd func()) (cron.EntryID, error) {
	return ct.ctab.AddJob(spec, cron.FuncJob(cmd))
}

func (ct *cronTab) schedule(sched cron.Schedule, cmd func()) cron.EntryID {
	return ct.ctab.Schedule(sched, cron.FuncJob(cmd))
}

func (ct *cronTab) remove(id cron.EntryID) {
	ct.ctab.Remove(id)
}

func (ct *cronTab) entries() []*entry {
	ents := ct.ctab.Entries()
	result := make([]*entry, 0, len(ents))
	for _, ent := range ents {
		result = append(result, &entry{
			ID:   ent.ID,
			Next: ent.Next,
			Prev: ent.Prev,
		})
	}

	return result
}

type entry struct {
	ID   cron.EntryID `json:"id"`
	Next time.Time    `json:"next"`
	Prev time.Time    `json:"prev"`
}
