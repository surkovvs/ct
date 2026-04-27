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
	statusesByGroup := make(map[string]map[string][]component.StageStatus)

	for _, cond := range a.storage.GetConditions() {
		unfinishedStages := cond.GetUnfinishedStages()
		if len(unfinishedStages) > 0 {
			if statusesByGroup[cond.Group] == nil {
				statusesByGroup[cond.Group] = map[string][]component.StageStatus{
					cond.Name: unfinishedStages,
				}
			} else {
				statusesByGroup[cond.Group][cond.Name] = unfinishedStages
			}
		}
	}

	report := strings.Builder{}
	for group, statusesByModule := range statusesByGroup {
		report.WriteString("group '" + group + "': ")
		for module, statuses := range statusesByModule {
			report.WriteString("module '" + module + "': ")
			strStages := make([]string, 0, len(statuses))
			for _, status := range statuses {
				strStages = append(strStages, "stage '"+string(status.Stage)+"': "+string(status.Status))
			}
			report.WriteString(strings.Join(strStages, ", ") + "; ")
		}
	}

	if report.Len() > 0 {
		a.logger.Error(`graceful shutdown timeout exeeded, got unfinished modules`,
			"application", a.name,
			"unfinished stages", report.String(),
		)
	} else {
		a.logger.Error(`graceful shutdown timeout exeeded`,
			"application", a.name)
	}
}
