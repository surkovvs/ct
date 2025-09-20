package component

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/surkovvs/ct/ctapp/zorro"
	"github.com/surkovvs/ct/ctifaces"
)

type (
	hcDynamic struct {
		mu           *sync.Mutex
		hcProcessing chan struct{}
		hcErr        error
	}

	Comp struct {
		status    zorro.Zorro
		object    any
		hc        *hcDynamic
		execWg    *sync.WaitGroup
		sdWgFlag  *uint32
		groupName string
		name      string
	}
)

type Define struct {
	GroupName string
	CompName  string
	Component any
	ExecWg    *sync.WaitGroup
}

func DefineComponent(d Define) Comp {
	status := zorro.New()
	if _, ok := d.Component.(ctifaces.Initializer); ok {
		d.ExecWg.Add(1)
		status.SetStatus(ready, initMask)
	}
	if _, ok := d.Component.(ctifaces.Runner); ok {
		d.ExecWg.Add(1)
		status.SetStatus(ready, runMask)
	}
	if _, ok := d.Component.(ctifaces.Shutdowner); ok {
		d.ExecWg.Add(1)
		status.SetStatus(ready, shutdownMask)
	}
	if _, ok := d.Component.(ctifaces.Healthchecker); ok {
		status.SetStatus(ready, healthcheckMask)
	}
	return Comp{
		execWg:    d.ExecWg,
		name:      d.CompName,
		object:    d.Component,
		status:    status,
		groupName: d.GroupName,
		hc:        &hcDynamic{mu: &sync.Mutex{}},
		sdWgFlag:  new(uint32),
	}
}

func (c Comp) IsValid() bool {
	return c.status.GetStatus() != 0
}

func (c Comp) GroupName() string {
	return c.groupName
}

func (c Comp) Name() string {
	return c.name
}

func (c Comp) GetCondition() Condition {
	stat := c.status.GetStatus()
	return Condition{
		Name:        c.name,
		Group:       c.groupName,
		Init:        namedStatuses[zorro.Status(stat.Querying(initMask)).ShiftTrailingZeros(initMask)],
		Run:         namedStatuses[zorro.Status(stat.Querying(runMask)).ShiftTrailingZeros(runMask)],
		Shutdown:    namedStatuses[zorro.Status(stat.Querying(shutdownMask)).ShiftTrailingZeros(shutdownMask)],
		Healthcheck: namedStatuses[zorro.Status(stat.Querying(healthcheckMask)).ShiftTrailingZeros(healthcheckMask)],
	}
}

func (c Comp) IsInitializer() bool {
	return c.status.GetStatus().Querying(initMask) != 0
}

func (c Comp) IsRunner() bool {
	return c.status.GetStatus().Querying(runMask) != 0
}

func (c Comp) IsShutdowner() bool {
	return c.status.GetStatus().Querying(shutdownMask) != 0
}

func (c Comp) IsHealthchecker() bool {
	return c.status.GetStatus().Querying(healthcheckMask) != 0
}

func (c Comp) genReport() Report {
	return Report{
		group:  c.groupName,
		module: c.name,
	}
}

func (c Comp) initializer() statusProvider {
	return statusProvider{
		provided: initMask,
		comp:     c,
	}
}

func (c Comp) runner() statusProvider {
	return statusProvider{
		provided: runMask,
		comp:     c,
	}
}

func (c Comp) shutdowner() statusProvider {
	return statusProvider{
		provided: shutdownMask,
		comp:     c,
	}
}

func (c Comp) healthchecker() statusProvider {
	return statusProvider{
		provided: healthcheckMask,
		comp:     c,
	}
}

func (r Comp) Init(ctx context.Context) Report {
	rep := r.genReport()
	switch {
	case !r.IsInitializer():
		rep.Code = CodeInfo
		rep.message = "no init method, skip"
		return rep
	case !r.initializer().isReady():
		rep.Code = CodeError
		rep.message = fmt.Sprintf("invalid status to start init (current status '%s')", r.initializer().namedStatus())
		rep.Err = ErrIncorrectStatusForAction
		return rep
	case ctx.Err() != nil:
		defer r.execWg.Done()
		rep.Code = CodeError
		rep.message = "context canceled for init, skip"
		rep.Err = ctx.Err()
		return rep
	}

	initializer, ok := r.object.(ctifaces.Initializer)
	if !ok {
		panic(fmt.Sprintf(`group '%s', module '%s', incorrectly defined as Initializer`, r.groupName, r.name))
	}

	r.initializer().setInProcess()
	defer r.execWg.Done()
	if err := initializer.Init(ctx); err != nil {
		r.initializer().setFailed()
		rep.Code = CodeError
		rep.message = "init error"
		rep.Err = err
		return rep
	}
	r.initializer().setDone()
	return rep
}

func (r Comp) Run(ctx context.Context) Report {
	rep := r.genReport()
	switch {
	case !r.IsRunner():
		rep.Code = CodeInfo
		rep.message = "no run method, skip"
		return rep
	case !r.runner().isReady():
		rep.Code = CodeError
		rep.message = fmt.Sprintf("invalid status to start run (current status '%s')", r.runner().namedStatus())
		rep.Err = ErrIncorrectStatusForAction
		return rep
	case r.IsInitializer() && !r.initializer().isDone():
		rep.Code = CodeError
		rep.message = fmt.Sprintf("trying to start run with init status '%s'", r.initializer().namedStatus())
		rep.Err = ErrIncorrectStatusForAction
		return rep
	case ctx.Err() != nil:
		defer r.execWg.Done()
		rep.Code = CodeError
		rep.message = "context canceled for run, skip"
		rep.Err = ctx.Err()
		return rep
	}

	runner, ok := r.object.(ctifaces.Runner)
	if !ok {
		panic(fmt.Sprintf(`group '%s', module '%s', incorrectly defined as Runner`, r.groupName, r.name))
	}

	r.runner().setInProcess()
	defer r.execWg.Done()
	if err := runner.Run(ctx); err != nil {
		r.runner().setFailed()
		rep.Code = CodeError
		rep.message = "run error"
		rep.Err = err
		return rep
	}
	r.runner().setDone()
	return rep
}

func (r Comp) Shutdown(ctx context.Context) Report {
	rep := r.genReport()
	switch {
	case !r.IsShutdowner():
		rep.Code = CodeInfo
		rep.message = "no shutdown method, skip"
		return rep
	case !r.IsRunner():
		defer func() {
			if atomic.CompareAndSwapUint32(r.sdWgFlag, 0, 1) {
				r.execWg.Done()
			}
		}()
		rep.Code = CodeInfo
		rep.message = "no run method, shutdown delayed"
		return rep
	}
	return r.ForseShutdown(ctx)
}

// ForseShutdown - starts regardless of the runner's status.
func (r Comp) ForseShutdown(ctx context.Context) Report {
	rep := r.genReport()
	switch {
	case !r.IsShutdowner():
		rep.Code = CodeInfo
		rep.message = "no shutdown method, skip"
		return rep
	case r.shutdowner().isDone() || r.shutdowner().isFailed() || r.shutdowner().isInProcess():
		rep.Code = CodeGSDrecall
		rep.message = fmt.Sprintf("re-calling a shutdown method (current status '%s'), skip",
			r.shutdowner().namedStatus())
		return rep
	case ctx.Err() != nil:
		defer func() {
			if atomic.CompareAndSwapUint32(r.sdWgFlag, 0, 1) {
				r.execWg.Done()
			}
		}()
		rep.Code = CodeError
		rep.message = "context canceled for shutdown, skip"
		rep.Err = ctx.Err()
		return rep
	}

	shutdowner, ok := r.object.(ctifaces.Shutdowner)
	if !ok {
		panic(fmt.Sprintf(`group '%s', module '%s', incorrectly defined as Shutdowner`, r.groupName, r.name))
	}

	r.shutdowner().setInProcess()
	defer func() {
		if atomic.CompareAndSwapUint32(r.sdWgFlag, 0, 1) {
			r.execWg.Done()
		}
	}()
	if err := shutdowner.Shutdown(ctx); err != nil {
		r.shutdowner().setFailed()
		rep.Code = CodeError
		rep.message = "shutdown error"
		rep.Err = err
		return rep
	}
	r.shutdowner().setDone()
	return rep
}

func (r Comp) Healthcheck(ctx context.Context) Report {
	rep := r.genReport()
	if !r.healthchecker().isInProcess() {
		r.hc.mu.Lock()
		defer r.hc.mu.Unlock()

		r.hc.hcProcessing = make(chan struct{})
		defer close(r.hc.hcProcessing)

		r.healthchecker().setInProcess()

		healthchecker, ok := r.object.(ctifaces.Healthchecker)
		if !ok {
			panic(fmt.Sprintf(`group '%s', module '%s', incorrectly defined as Healthchecker`, r.groupName, r.name))
		}

		if err := healthchecker.Healthcheck(ctx); err != nil {
			r.healthchecker().setFailed()
			r.hc.hcErr = err

			rep.Code = CodeError
			rep.message = "healthcheck error"
			rep.Err = err
			return rep
		}
		r.healthchecker().setDone()
		return rep
	}

	<-r.hc.hcProcessing
	if r.hc.hcErr != nil {
		rep.Code = CodeError
		rep.message = "healthcheck error"
		rep.Err = r.hc.hcErr
	}
	return rep
}
