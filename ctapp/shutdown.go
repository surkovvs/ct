package ctapp

import (
	"strings"
	"time"

	"github.com/surkovvs/ct/ctapp/component"
)

func (a *App) gracefulShutdown() {
	defer close(a.shutdown.shutdownDone)

	go func() {
		time.Sleep(*a.shutdown.timeout)
		a.shutdown.ctxCancel()
	}()

	select {
	case <-a.shutdown.ctx.Done():
		a.reportUnfinished()
	case <-a.execution.execDone:
		select {
		case <-a.shutdown.ctx.Done():
			a.reportUnfinished()
		default:
			a.logger.Info(`graceful shutdown finished`,
				"application", a.name)
		}
	}
}

func (a *App) reportUnfinished() {
	ufInit := make(map[string][]string)
	ufRun := make(map[string][]string)
	ufShutdown := make(map[string][]string)
	ufHealthcheck := make(map[string][]string)

	for _, cond := range a.storage.GetConditions() {
		if cond.Init == component.InProcessString {
			ufInit[cond.Group] = append(ufInit[cond.Group], cond.Name)
		}
		if cond.Run == component.InProcessString {
			ufRun[cond.Group] = append(ufRun[cond.Group], cond.Name)
		}
		if cond.Shutdown == component.InProcessString {
			ufShutdown[cond.Group] = append(ufShutdown[cond.Group], cond.Name)
		}
		if cond.Healthcheck == component.InProcessString {
			ufHealthcheck[cond.Group] = append(ufHealthcheck[cond.Group], cond.Name)
		}
	}

	var args []any
	if len(ufInit) > 0 {
		args = append(args, "initialization", modulesByGroupsToString(ufInit))
	}
	if len(ufRun) > 0 {
		args = append(args, "running", modulesByGroupsToString(ufRun))
	}
	if len(ufShutdown) > 0 {
		args = append(args, "shutdown", modulesByGroupsToString(ufShutdown))
	}
	if len(ufHealthcheck) > 0 {
		args = append(args, "healthcheck", modulesByGroupsToString(ufHealthcheck))
	}

	if len(args) > 0 {
		args = append([]any{"application", a.name}, args...)
		a.logger.Error(`graceful shutdown timeout exeeded, got unfinished modules`,
			args...,
		)
	} else {
		a.logger.Error(`graceful shutdown timeout exeeded`,
			"application", a.name)
	}
}

func modulesByGroupsToString(m map[string][]string) string {
	byGroups := make([]string, 0, len(m))
	for group, modules := range m {
		byGroups = append(byGroups, "group '"+group+"': "+strings.Join(modules, ", "))
	}

	return strings.Join(byGroups, ";")
}
