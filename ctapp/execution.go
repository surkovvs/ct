package ctapp

import (
	"context"
	"errors"
	"sync"

	"github.com/surkovvs/ct/ctapp/component"
	"github.com/surkovvs/ct/ctapp/compstor"
	"github.com/surkovvs/ct/ctapp/vector"
	"github.com/surkovvs/ct/ctapp/wgchan"
)

func (a *App) exec() {
	var vecs []any
	if vec := a.buildEgrInitVector(a.execution.initCtx); vec != nil {
		vecs = append(vecs, vec)
	}
	if vec := a.buildSeqInitVec(a.execution.initCtx); vec != nil {
		vecs = append(vecs, vec)
	}
	if vec := a.buildIngrInitVec(a.execution.initCtx); vec != nil {
		vecs = append(vecs, vec)
	}
	vec, wg := a.buildRunVec(a.execution.runCtx)
	vecs = append(vecs, vec)
	go func() {
		<-wgchan.NewWgChan(wg)
		close(a.execution.runDone)
	}()

	if vec := a.buildIngrSdVec(a.shutdown.ctx); vec != nil {
		vecs = append(vecs, vec)
	}
	if vec := a.buildSeqSdVec(a.shutdown.ctx); vec != nil {
		vecs = append(vecs, vec)
	}
	if vec := a.buildEgrSdVec(a.shutdown.ctx); vec != nil {
		vecs = append(vecs, vec)
	}

	c := vector.NewConstructor(a.execution.reports)
	runWithoutCancel := context.WithoutCancel(a.execution.runCtx)
	c.Sequentially(runWithoutCancel, vecs...).Exec(runWithoutCancel)
	close(a.execution.execDone)
}

// need to be checked on nil result
func (a *App) buildEgrInitVector(vecCtx context.Context) *vector.Vector[component.Report] {
	egrGroup, err := a.storage.GetGroupByName(EgressGroup)
	if err != nil {
		if errors.Is(err, compstor.ErrGroupNotFound) {
			a.logger.Debug(`egress group not found`,
				"application", a.name)
			return nil
		} else {
			a.logger.Error(`unexpected error`,
				"application", a.name,
				`group`, egrGroup.GetName(),
				"error", err)
		}
	}
	var egrInits []any
	for _, module := range egrGroup.GetComponents() {
		egrInits = append(egrInits, module.Init)
	}
	c := vector.NewConstructor(a.execution.reports)
	return c.Concurrently(vecCtx, egrInits...)
}

// need to be checked on nil result
func (a *App) buildSeqInitVec(vecCtx context.Context) *vector.Vector[component.Report] {
	groupList := a.storage.GetOrderedGroupList()
	seqGroupVecs := make([]any, 0, len(groupList))
	c := vector.NewConstructor(a.execution.reports)
	for _, group := range groupList {
		if group.GetName() == IngressGroup || group.GetName() == EgressGroup {
			continue
		}
		var seqGroupInits []any
		for _, module := range group.GetComponents() {
			seqGroupInits = append(seqGroupInits, module.Init)
		}
		seqGroupVecs = append(seqGroupVecs, c.Sequentially(vecCtx, seqGroupInits...))
	}
	if len(seqGroupVecs) == 0 {
		return nil
	}
	return c.Concurrently(vecCtx, seqGroupVecs...)
}

// need to be checked on nil result
func (a *App) buildIngrInitVec(vecCtx context.Context) *vector.Vector[component.Report] {
	ingrGroup, err := a.storage.GetGroupByName(IngressGroup)
	if err != nil {
		if errors.Is(err, compstor.ErrGroupNotFound) {
			a.logger.Debug(`egress group not found`,
				"application", a.name)
			return nil
		} else {
			a.logger.Error(`unexpected error`,
				"application", a.name,
				`group`, ingrGroup.GetName(),
				"error", err)
		}
	}
	var ingrInits []any
	for _, module := range ingrGroup.GetComponents() {
		ingrInits = append(ingrInits, module.Init)
	}
	c := vector.NewConstructor(a.execution.reports)
	return c.Concurrently(vecCtx, ingrInits...)
}

// at least 1 module must be added
func (a *App) buildRunVec(vecCtx context.Context) (*vector.Vector[component.Report], *sync.WaitGroup) {
	groupList := a.storage.GetOrderedGroupList()
	groupVecs := make([]any, 0, len(groupList))
	c := vector.NewConstructor(a.execution.reports)
	for _, group := range groupList {
		var groupVec *vector.Vector[component.Report]
		if group.GetName() == IngressGroup || group.GetName() == EgressGroup {
			var concGroupRuns []any
			for _, module := range group.GetComponents() {
				concGroupRuns = append(concGroupRuns, module.Run)
			}
			groupVec = c.Concurrently(vecCtx, concGroupRuns...)
		} else {
			var seqGroupRuns []any
			for _, module := range group.GetComponents() {
				seqGroupRuns = append(seqGroupRuns, module.Run)
			}
			groupVec = c.Sequentially(vecCtx, seqGroupRuns...)
		}
		groupVecs = append(groupVecs, groupVec)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	return c.WithReleaseWG(&wg, c.Concurrently(vecCtx, groupVecs...)), &wg
}

// need to be checked on nil result
func (a *App) buildIngrSdVec(vecCtx context.Context) *vector.Vector[component.Report] {
	ingrGroup, err := a.storage.GetGroupByName(IngressGroup)
	if err != nil {
		if errors.Is(err, compstor.ErrGroupNotFound) {
			a.logger.Debug(`egress group not found`,
				"application", a.name)
			return nil
		} else {
			a.logger.Error(`unexpected error`,
				"application", a.name,
				`group`, ingrGroup.GetName(),
				"error", err)
		}
	}
	var ingrShutdowns []any
	for _, module := range ingrGroup.GetComponents() {
		ingrShutdowns = append(ingrShutdowns, module.Shutdown)
	}
	c := vector.NewConstructor(a.execution.reports)
	return c.Concurrently(vecCtx, ingrShutdowns...)
}

// need to be checked on nil result
func (a *App) buildSeqSdVec(vecCtx context.Context) *vector.Vector[component.Report] {
	groupList := a.storage.GetOrderedGroupList()
	seqGroupVecs := make([]any, 0, len(groupList))
	c := vector.NewConstructor(a.execution.reports)
	for _, group := range groupList {
		if group.GetName() == IngressGroup || group.GetName() == EgressGroup {
			continue
		}
		var seqGroupShutdowns []any
		for _, module := range group.GetComponents() {
			seqGroupShutdowns = append(seqGroupShutdowns, module.Shutdown)
		}
		seqGroupVecs = append(seqGroupVecs, c.Sequentially(vecCtx, seqGroupShutdowns...))
	}
	if len(seqGroupVecs) == 0 {
		return nil
	}
	return c.Concurrently(vecCtx, seqGroupVecs...)
}

// need to be checked on nil result
func (a *App) buildEgrSdVec(vecCtx context.Context) *vector.Vector[component.Report] {
	egrGroup, err := a.storage.GetGroupByName(EgressGroup)
	if err != nil {
		if errors.Is(err, compstor.ErrGroupNotFound) {
			a.logger.Debug(`egress group not found`,
				"application", a.name)
			return nil
		} else {
			a.logger.Error(`unexpected error`,
				"application", a.name,
				`group`, egrGroup.GetName(),
				"error", err)
		}
	}
	var egrShutdowns []any
	for _, module := range egrGroup.GetComponents() {
		egrShutdowns = append(egrShutdowns, module.Shutdown)
	}
	c := vector.NewConstructor(a.execution.reports)
	return c.Concurrently(vecCtx, egrShutdowns...)
}
