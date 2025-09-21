package ctapp

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/surkovvs/ct/ctapp/component"
	"github.com/surkovvs/ct/ctapp/compstor"
	"github.com/surkovvs/ct/ctifaces"
)

//nolint:gochecknoglobals // as planned
var (
	DefaultProvidedSigs    = []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	DefaultShutdownTimeout = time.Second * 3
	BackgroundGroup        = `background`      // Concurrent running group.
	BackgroundSyncGroup    = `background_sync` // Group, where initialization of modules is synced.
)

type (
	execution struct {
		wg            *sync.WaitGroup
		done          chan struct{}
		reports       chan component.Report
		initCtx       context.Context
		runCtx        context.Context
		initRunCancel context.CancelFunc
		initTimeout   *time.Duration
		tolerantMode  bool
	}
	shutdown struct {
		ctx          context.Context
		ctxCancel    context.CancelFunc
		shutdownDone chan struct{}
		sigs         []os.Signal
		timeout      *time.Duration
		exitCode     int
	}
	App struct {
		execution execution
		shutdown  shutdown
		storage   compstor.CompsStorage
		name      string
		logger    ctifaces.Logger
	}
)

func New(opts ...AppOption) *App {
	sdCtx, sdCancel := context.WithCancel(context.Background())
	a := &App{
		execution: execution{
			wg:            &sync.WaitGroup{},
			done:          make(chan struct{}),
			reports:       make(chan component.Report),
			initCtx:       nil,
			runCtx:        nil,
			initRunCancel: nil,
			initTimeout:   nil,
			tolerantMode:  false,
		},
		shutdown: shutdown{
			ctx:          sdCtx,
			ctxCancel:    sdCancel,
			shutdownDone: make(chan struct{}),
			sigs:         nil,
			timeout:      nil,
			exitCode:     0,
		},
		storage: compstor.NewCompsStorage(),
		name:    "",
		logger:  nil,
	}

	for _, opt := range opts {
		opt(a)
	}
	a.defaultSettingsCheckAndApply()

	if err := a.storage.AddGroup(BackgroundGroup); err != nil &&
		!errors.Is(err, compstor.ErrGroupAlreadyRegistered) {
		a.logger.Error(`group addition`,
			"application", a.name,
			`group`, BackgroundGroup,
			`error`, err)
		os.Exit(1)
	}

	return a
}

func (a *App) accompaniment() {
	syscallC := make(chan os.Signal, 1)
	signal.Notify(syscallC, a.shutdown.sigs...)
	execDone := a.execution.done
CycleLable:
	for {
		select {
		case rep := <-a.execution.reports:
			switch rep.Code {
			case component.CodeEmpty:
			case component.CodeInfo, component.CodeGSDrecall:
				a.logger.Debug(`module message`,
					`application`, a.name,
					`message`, rep.String())
			case component.CodeError:
				a.logger.Error(`module error`,
					"application", a.name,
					`error`, rep.String())
				if !a.execution.tolerantMode {
					execDone = nil
					signal.Stop(syscallC)

					a.logger.Debug(`execution failed graceful shutdown started`,
						"application", a.name)

					a.execution.initRunCancel()
					go a.gracefulShutdown()

					a.execution.tolerantMode = true
				}
			}
		case <-execDone:
			execDone = nil
			signal.Stop(syscallC)

			a.logger.Debug(`execution finished graceful shutdown started`,
				"application", a.name)

			a.execution.initRunCancel()
			go a.gracefulShutdown()
		case sig := <-syscallC:
			execDone = nil
			signal.Stop(syscallC)

			a.logger.Info(`graceful shutdown started by syscall`,
				"application", a.name,
				`syscall`, sig.String())

			a.execution.initRunCancel()
			go a.gracefulShutdown()
		case <-a.shutdown.shutdownDone:
			break CycleLable
		}
	}
}
