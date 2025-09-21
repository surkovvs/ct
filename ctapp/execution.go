package ctapp

import (
	"errors"
	"sync"

	"github.com/surkovvs/ct/ctapp/compstor"
	"github.com/surkovvs/ct/ctapp/vector"
)

func (a *App) exec() {
	// backgroung groups runs before others, all the components in background group runs concurrently
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.processBackground()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.processBackgroundSync()
	}()
	wg.Wait()

	a.processSequentialGroups()

	a.execution.wg.Wait()
	close(a.execution.done)
}

func (a *App) processBackground() {
	bgGroup, err := a.storage.GetGroupByName(BackgroundGroup)
	if err != nil {
		if errors.Is(err, compstor.ErrGroupNotFound) {
			a.logger.Debug(`background group not found`,
				"application", a.name)
		} else {
			a.logger.Error(`unexpected error`,
				"application", a.name,
				`group`, bgGroup.GetName(),
				"error", err)
		}
	} else {
		wg := sync.WaitGroup{}
		c := vector.NewConstructor(a.execution.reports)
		var targets []any
		for _, module := range bgGroup.GetComponents() {
			wg.Add(1)
			targets = append(targets,
				c.Sequentially(a.execution.runCtx,
					c.WithReleaseWG(&wg, c.Sequentially(a.execution.initCtx, module.Init)),
					c.Sequentially(a.execution.runCtx, module.Run),
					c.Sequentially(a.shutdown.ctx, module.Shutdown),
				),
			)
		}
		go c.Concurrently(a.execution.runCtx, targets...).Exec(a.execution.runCtx)

		wg.Wait()
	}
}

func (a *App) processBackgroundSync() {
	bgsGroup, err := a.storage.GetGroupByName(BackgroundSyncGroup)
	if err != nil {
		if errors.Is(err, compstor.ErrGroupNotFound) {
			a.logger.Debug(`background sync group not found`,
				"application", a.name)
		} else {
			a.logger.Error(`unexpected error`,
				"application", a.name,
				`group`, bgsGroup.GetName(),
				"error", err)
		}
	} else {
		wg := sync.WaitGroup{}
		c := vector.NewConstructor(a.execution.reports)
		var inits, rsd []any
		for _, module := range bgsGroup.GetComponents() {
			wg.Add(1)
			inits = append(inits, c.WithReleaseWG(&wg, c.Sequentially(a.execution.initCtx, module.Init)))
			rsd = append(rsd,
				c.Sequentially(a.execution.runCtx, module.Run,
					c.Sequentially(a.shutdown.ctx, module.Shutdown),
				),
			)
		}
		go c.Sequentially(a.execution.runCtx,
			c.Concurrently(a.execution.runCtx, inits...),
			c.Concurrently(a.execution.runCtx, rsd...),
		).Exec(a.execution.runCtx)

		wg.Wait()
	}
}

func (a *App) processSequentialGroups() {
	c := vector.NewConstructor(a.execution.reports)
	groupList := a.storage.GetOrderedGroupList()
	vectors := make([]any, 0, len(groupList))
	for _, group := range groupList {
		if group.GetName() == BackgroundGroup || group.GetName() == BackgroundSyncGroup {
			continue
		}
		var inits, runs, sds []any
		for _, module := range group.GetComponents() {
			inits = append(inits, module.Init)
			runs = append(runs, module.Run)
			sds = append(sds, module.Shutdown)
		}
		vectors = append(vectors, c.Sequentially(a.execution.runCtx,
			c.Sequentially(a.execution.initCtx, inits...),
			c.Sequentially(a.execution.runCtx, runs...),
			c.Sequentially(a.shutdown.ctx, sds...),
		))
	}
	go c.Concurrently(a.execution.runCtx, vectors...).Exec(a.execution.runCtx)
}
