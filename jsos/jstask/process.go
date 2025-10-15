package jstask

import (
	"time"

	"github.com/xmx/aegis-common/jsos/jsvm"
)

type Processor interface {
	PID() uint64
	Name() string
	Engineer() jsvm.Engineer
	StartedAt() time.Time
}

type process struct {
	pid       uint64
	name      string
	eng       jsvm.Engineer
	startedAt time.Time
	err       error
}

func (p *process) PID() uint64 {
	return p.pid
}

func (p *process) Name() string {
	return p.name
}

func (p *process) Engineer() jsvm.Engineer {
	return p.eng
}

func (p *process) StartedAt() time.Time {
	return p.startedAt
}

func (p *process) run(name, code string) error {
	p.startedAt = time.Now()
	_, err := p.eng.RunScript(name, code)
	return err
}

func (p *process) kill(cause any) {
	p.eng.Kill(cause)
}
