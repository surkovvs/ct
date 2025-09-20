package ctapp

import (
	"context"
	"sync"

	"github.com/surkovvs/ct/ctapp/component"
)

func (a *App) Healthcheck(ctx context.Context) []error {
	var errs []error
	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, module := range a.storage.GetUnsortedHealthcheckers() {
		wg.Add(1)
		go func(module component.Comp) {
			defer wg.Done()
			if rep := module.Healthcheck(ctx); rep.Code == component.CodeError {
				mu.Lock()
				errs = append(errs, rep.Err)
				mu.Unlock()
			}
		}(module)
	}
	wg.Wait()

	return errs
}
