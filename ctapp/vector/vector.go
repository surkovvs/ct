package vector

import (
	"context"
	"sync"

	"github.com/surkovvs/ct/ctapp/wgchan"
)

type basicFunc[T any] func(context.Context) T

type Vector[T any] struct {
	job    basicFunc[T]
	jobCtx context.Context
	conc   []*Vector[T]
	seq    *Vector[T]
	wait   *sync.WaitGroup
	done   *sync.WaitGroup

	report chan<- T
}

// Exec - blocking call.
func (v *Vector[T]) Exec(ctx context.Context) {
	if v.wait != nil {
		v.wait.Wait() // блокируемся, если нужно
	}

	// исполняем свою функцию
	if v.job != nil {
		jobDone := make(chan struct{})
		go func() {
			defer close(jobDone)
			v.report <- v.job(v.jobCtx)
		}()
		// ожидаем завершения джобы, либо закрытия контекста исполнения
		select {
		case <-ctx.Done():
		case <-jobDone:
		}
	}

	// конкурентно запускаем векторы, и ждем исполнения
	if len(v.conc) != 0 {
		wg := sync.WaitGroup{}
		for _, cVector := range v.conc {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cVector.Exec(ctx)
			}()
		}
		wg.Wait()
		select {
		case <-ctx.Done():
		case <-wgchan.NewWgChan(&wg):
		}
	}

	// высвобождаем блокировку, если можем
	if v.done != nil {
		v.done.Done()
		v.done = nil
	}

	// запускаем следующего
	if v.seq != nil {
		v.seq.Exec(ctx)
	}
}

type Constructor[T any] struct {
	report chan<- T
}

func NewConstructor[T any](report chan<- T) Constructor[T] {
	return Constructor[T]{
		report: report,
	}
}

func (c Constructor[T]) newVector() *Vector[T] {
	return &Vector[T]{
		report: c.report,
	}
}

func (c Constructor[T]) Sequentially(ctx context.Context, targets ...any) *Vector[T] {
	var head, tail *Vector[T]

	for i := len(targets) - 1; i >= 0; i-- {
		switch target := targets[i].(type) {
		case *Vector[T]:
			if target == nil {
				continue
			}
			if target.job != nil && target.jobCtx == nil {
				target.jobCtx = ctx
			}
			head = target
		case func(context.Context) T:
			head = c.newVector()
			head.job = target
			head.jobCtx = ctx
		default:
			panic("incorrect type passed to *vector.Consructor method")
		}

		if head.seq == nil {
			head.seq = tail
		} else {
			b := head
			for b.seq != nil {
				b = b.seq
			}
			b.seq = tail
		}

		tail = head
	}
	return head
}

func (c Constructor[T]) Concurrently(ctx context.Context, targets ...any) *Vector[T] {
	head := c.newVector()
	for _, target := range targets {
		switch target := target.(type) {
		case *Vector[T]:
			if target.job != nil && target.jobCtx == nil {
				target.jobCtx = ctx
			}
			head.conc = append(head.conc, target)
		case func(context.Context) T:
			sub := c.newVector()
			sub.jobCtx = ctx
			sub.job = target
			head.conc = append(head.conc, sub)
		default:
			panic("incorrect type passed to *vector.Consructor method")
		}
	}
	return head
}

func (c Constructor[T]) WithReleaseWG(wg *sync.WaitGroup, target any) *Vector[T] {
	var vector *Vector[T]
	switch target := target.(type) {
	case *Vector[T]:
		vector = target
	case func(context.Context) T:
		vector = c.newVector()
		vector.job = target
	default:
		panic("incorrect type passed to *vector.Consructor method")
	}
	vector.done = wg

	return vector
}

func (c Constructor[T]) WithWaitWG(wg *sync.WaitGroup, target any) *Vector[T] {
	var vector *Vector[T]
	switch target := target.(type) {
	case *Vector[T]:
		vector = target
	case func(context.Context) T:
		vector = c.newVector()
		vector.job = target
	default:
		panic("incorrect type passed to *vector.Consructor method")
	}
	vector.wait = wg

	return vector
}
