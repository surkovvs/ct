package component

import (
	"context"
	"fmt"
	"sync"

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
		groupName string
		name      string
	}
)

type Define struct {
	GroupName string
	CompName  string
	Component any
}

func DefineComponent(d Define) Comp {
	status := zorro.New()
	if _, ok := d.Component.(ctifaces.Initializer); ok {
		status.SetStatus(ready, initMask)
	}
	if _, ok := d.Component.(ctifaces.Runner); ok {
		status.SetStatus(ready, runMask)
	}
	if _, ok := d.Component.(ctifaces.Shutdowner); ok {
		status.SetStatus(ready, shutdownMask)
	}
	if _, ok := d.Component.(ctifaces.Healthchecker); ok {
		status.SetStatus(ready, healthcheckMask)
	}
	return Comp{
		name:      d.CompName,
		object:    d.Component,
		status:    status,
		groupName: d.GroupName,
		hc: &hcDynamic{
			mu:           &sync.Mutex{},
			hcProcessing: nil,
			hcErr:        nil,
		},
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
		group:   c.groupName,
		module:  c.name,
		Err:     nil,
		message: "",
		Code:    0,
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

func (c Comp) Init(ctx context.Context) Report {
	rep := c.genReport()
	switch {
	case !c.IsInitializer():
		rep.Code = CodeInfo
		rep.message = "no init method, skip"
		return rep
	case ctx.Err() != nil:
		rep.Code = CodeError
		rep.message = "context canceled for init, skip"
		rep.Err = ctx.Err()
		return rep
	}

	initializer, ok := c.object.(ctifaces.Initializer)
	if !ok {
		panic(fmt.Sprintf(`group '%s', module '%s', incorrectly defined as Initializer`, c.groupName, c.name))
	}

	if !c.initializer().tryChangeStatus(ready, inProcess) {
		rep.Code = CodeError
		rep.message = fmt.Sprintf("invalid status to start init (current status '%s')", c.initializer().namedStatus())
		rep.Err = ErrIncorrectStatusForAction
		return rep
	}
	if err := initializer.Init(ctx); err != nil {
		c.initializer().setFailed()
		rep.Code = CodeError
		rep.message = "init error"
		rep.Err = err
		return rep
	}
	c.initializer().setDone()
	return rep
}

func (c Comp) Run(ctx context.Context) Report {
	rep := c.genReport()
	switch {
	case !c.IsRunner():
		rep.Code = CodeInfo
		rep.message = "no run method, skip"
		return rep
	case c.IsInitializer() && !c.initializer().isDone():
		rep.Code = CodeError
		rep.message = fmt.Sprintf("trying to start run with init status '%s'", c.initializer().namedStatus())
		rep.Err = ErrIncorrectStatusForAction
		return rep
	case ctx.Err() != nil:
		rep.Code = CodeError
		rep.message = "context canceled for run, skip"
		rep.Err = ctx.Err()
		return rep
	}

	runner, ok := c.object.(ctifaces.Runner)
	if !ok {
		panic(fmt.Sprintf(`group '%s', module '%s', incorrectly defined as Runner`, c.groupName, c.name))
	}

	if !c.runner().tryChangeStatus(ready, inProcess) {
		rep.Code = CodeError
		rep.message = fmt.Sprintf("invalid status to start run (current status '%s')", c.runner().namedStatus())
		rep.Err = ErrIncorrectStatusForAction
		return rep
	}
	if err := runner.Run(ctx); err != nil {
		c.runner().setFailed()
		rep.Code = CodeError
		rep.message = "run error"
		rep.Err = err
		return rep
	}
	c.runner().setDone()
	return rep
}

func (c Comp) Shutdown(ctx context.Context) Report {
	rep := c.genReport()
	switch {
	case !c.IsShutdowner():
		rep.Code = CodeInfo
		rep.message = "no shutdown method, skip"
		return rep
	case !c.IsRunner():
		rep.Code = CodeInfo
		rep.message = "no run method, shutdown delayed"
		return rep
	}
	return c.ForseShutdown(ctx)
}

// ForseShutdown - starts regardless of the runner's status.
func (c Comp) ForseShutdown(ctx context.Context) Report {
	rep := c.genReport()
	switch {
	case !c.IsShutdowner():
		rep.Code = CodeInfo
		rep.message = "no shutdown method, skip"
		return rep
	case ctx.Err() != nil:
		rep.Code = CodeError
		rep.message = "context canceled for shutdown, skip"
		rep.Err = ctx.Err()
		return rep
	}

	shutdowner, ok := c.object.(ctifaces.Shutdowner)
	if !ok {
		panic(fmt.Sprintf(`group '%s', module '%s', incorrectly defined as Shutdowner`, c.groupName, c.name))
	}

	if !c.shutdowner().tryChangeStatus(ready, inProcess) {
		rep.Code = CodeGSDrecall
		rep.message = fmt.Sprintf("re-calling a shutdown method (current status '%s'), skip",
			c.shutdowner().namedStatus())
		return rep
	}
	if err := shutdowner.Shutdown(ctx); err != nil {
		c.shutdowner().setFailed()
		rep.Code = CodeError
		rep.message = "shutdown error"
		rep.Err = err
		return rep
	}
	c.shutdowner().setDone()
	return rep
}

func (c Comp) Healthcheck(ctx context.Context) Report {
	rep := c.genReport()
	if !c.healthchecker().isInProcess() {
		c.hc.mu.Lock()
		defer c.hc.mu.Unlock()

		c.hc.hcProcessing = make(chan struct{})
		defer close(c.hc.hcProcessing)

		c.healthchecker().setInProcess()

		healthchecker, ok := c.object.(ctifaces.Healthchecker)
		if !ok {
			panic(fmt.Sprintf(`group '%s', module '%s', incorrectly defined as Healthchecker`, c.groupName, c.name))
		}

		if err := healthchecker.Healthcheck(ctx); err != nil {
			c.healthchecker().setFailed()
			c.hc.hcErr = err

			rep.Code = CodeError
			rep.message = "healthcheck error"
			rep.Err = err
			return rep
		}
		c.healthchecker().setDone()
		return rep
	}

	<-c.hc.hcProcessing
	if c.hc.hcErr != nil {
		rep.Code = CodeError
		rep.message = "healthcheck error"
		rep.Err = c.hc.hcErr
	}
	return rep
}
