package jsmod

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/grafana/sobek"
	"github.com/xmx/aegis-common/jsos/jsvm"
	"github.com/xmx/aegis-common/library/cronv3"
)

func NewCrontab(crond *cronv3.Crontab) jsvm.Module {
	return &crontabModule{crond: crond}
}

type crontabModule struct {
	eng   jsvm.Engineer
	crond *cronv3.Crontab
}

func (mod *crontabModule) Preload(eng jsvm.Engineer) (string, any, bool) {
	mod.eng = eng
	vals := map[string]any{
		"addJob": mod.addJob,
	}

	return "crontab", vals, false
}

func (mod *crontabModule) addJob(spec string, cmd func()) (*sobek.Object, error) {
	buf := make([]byte, 24)
	nano := time.Now().UnixNano()
	binary.BigEndian.PutUint64(buf, uint64(nano))
	_, _ = rand.Read(buf[8:])
	id := "js-crontab-" + hex.EncodeToString(buf)
	if _, err := mod.crond.AddJob(id, spec, cmd); err != nil {
		return nil, err
	}

	rt := mod.eng.Runtime()
	parent := mod.eng.Context()
	finalize := mod.eng.Finalizer()
	ctx, cancel := context.WithCancel(parent)
	handle := &taskHandle{
		cid:    id,
		crond:  mod.crond,
		final:  finalize,
		ctx:    ctx,
		cancel: cancel,
	}
	handle.fid = finalize.Add(handle.remove)

	ret := rt.NewObject()
	_ = ret.Set("id", handle.id)
	_ = ret.Set("wait", handle.wait)
	_ = ret.Set("remove", handle.remove)

	return ret, nil
}

type taskHandle struct {
	cid    string
	fid    uint64
	final  jsvm.Finalizer
	crond  *cronv3.Crontab
	ctx    context.Context
	cancel context.CancelFunc
}

func (h *taskHandle) id() string {
	return h.cid
}

func (h *taskHandle) wait() {
	<-h.ctx.Done()
}

func (h *taskHandle) remove() {
	h.crond.Remove(h.cid)
	h.final.Del(h.fid)
	h.cancel()
}
