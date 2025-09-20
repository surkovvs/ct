//nolint:mnd // example
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/surkovvs/ct/ctapp/vector"
)

const (
	init3Duration     = time.Millisecond * 250
	run1Duration      = time.Millisecond * 500
	shutdown1Duration = time.Millisecond * 500

	releaseShutdown2Since = time.Millisecond * 1250
	releaseGroup3Since    = time.Millisecond * 1000

	cancelExecCtxSince   = time.Millisecond * 1500
	cancelGroup1CtxSince = time.Millisecond * 750
)

func main() {
	start := time.Now()

	// log.SetFlags(log.Lmicroseconds)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "since_start" {
				a.Value = slog.DurationValue(a.Value.Time().Sub(start))
			}
			return a
		},
	}),
	))

	execCtx, cancelExecCtx := context.WithCancel(context.Background())
	group1Ctx, cancelGroup1Ctx := context.WithCancel(context.Background())
	ctx := context.Background()

	stubWaiting := sync.WaitGroup{}
	stubWaiting.Add(1)
	initWaiting := sync.WaitGroup{}
	initWaiting.Add(1)
	blockShutdown2 := sync.WaitGroup{}
	blockShutdown2.Add(1)
	blockGroup3 := sync.WaitGroup{}
	blockGroup3.Add(1)

	wg := sync.WaitGroup{}
	wg.Add(3)

	c := vector.NewConstructor(logReceiver())
	v := c.Sequentially(ctx,
		c.WithReleaseWG(&initWaiting,
			c.Concurrently(ctx,
				stub("init_1", 0), stub("init_2", 0), stub("init_3", init3Duration),
			),
		),
		c.Concurrently(ctx,
			c.Sequentially(group1Ctx,
				stub("run_1", run1Duration),
				c.WithReleaseWG(&wg,
					stub("shutdown_1", shutdown1Duration),
				),
			),
			c.Sequentially(ctx,
				stub("run_2", 0),
				c.WithReleaseWG(&wg,
					c.WithWaitWG(&blockShutdown2,
						stub("shutdown_2", 0),
					)),
			),
			c.WithWaitWG(&blockGroup3,
				c.Sequentially(ctx,
					stub("run_3", 0),
					c.WithReleaseWG(&wg,
						stub("shutdown_3", 0)),
				),
			),
		),
		c.WithReleaseWG(&stubWaiting, stub("stub", 0)),
	)

	fmt.Printf("vector builded since: %v\n", time.Since(start))

	go func() {
		time.Sleep(releaseShutdown2Since)
		blockShutdown2.Done()
	}()

	go func() {
		time.Sleep(releaseGroup3Since)
		blockGroup3.Done()
	}()

	go func() {
		time.Sleep(cancelExecCtxSince)
		cancelExecCtx()
	}()

	go func() {
		time.Sleep(cancelGroup1CtxSince)
		cancelGroup1Ctx()
	}()

	execution := time.Now()

	go func() {
		stubWaiting.Wait()
		fmt.Printf("stubWaiting released since vector build: %v\n", time.Since(execution))
	}()

	go func() {
		initWaiting.Wait()
		fmt.Printf("initWaiting released since vector build: %v\n", time.Since(execution))
	}()

	v.Exec(execCtx)
	defer time.Sleep(time.Millisecond * 2000)

	fmt.Printf("all shutdowns ended since: %v\n", time.Since(execution))
	fmt.Printf("total run time since start: %v\n", time.Since(start))
}

func logReceiver() chan<- error {
	errChan := make(chan error)
	go func() {
		for err := range errChan {
			log.Println("err", err)
		}
	}()

	return errChan
}

func stub(tag string, dur time.Duration) func(context.Context) error {
	return func(ctx context.Context) error {
		if dur != 0 {
			time.Sleep(dur)
		}
		if ctx.Err() != nil {
			return errors.New(tag + " ctx canceled")
		}
		slog.Info("stub report", "since_start", time.Now(), "tag", tag, "status", "finished")
		return nil
	}
}
