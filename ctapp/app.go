package ctapp

import (
	"context"
	"os"
	"os/signal"
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

	IngressGroup = `ingress`
	EgressGroup  = `egress`
)

type (
	execution struct {
		runDone       chan struct{} // reports, jobs running are finished (excluding shutdowns)
		execDone      chan struct{} // reports, that all the jobs are finished (including shutdowns)
		reports       chan component.Report
		initCtx       context.Context // based on runCtx, separated for init timeout
		runCtx        context.Context // main context in app
		initRunCancel context.CancelFunc
		initTimeout   *time.Duration
		tolerantMode  bool // forced switch in case of interrupt
	}
	shutdown struct {
		ctx          context.Context
		ctxCancel    context.CancelFunc // cancelation, based on sd timeout, processed in gracefulShutdown
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
			runDone:       make(chan struct{}),
			execDone:      make(chan struct{}),
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

	return a
}

func (a *App) accompaniment() {
	syscallC := make(chan os.Signal, 1)
	signal.Notify(syscallC, a.shutdown.sigs...)
	execDone := a.execution.execDone
	runDone := a.execution.runDone
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
					runDone = nil
					signal.Stop(syscallC)

					a.logger.Debug(`execution failed graceful shutdown started`,
						"application", a.name)

					a.execution.initRunCancel()
					go a.gracefulShutdown()
					a.execution.tolerantMode = true
				}
			}
		case <-runDone:
			runDone = nil
			signal.Stop(syscallC)

			a.logger.Debug(`all runs are finished, graceful shutdown started`,
				"application", a.name)

			a.execution.initRunCancel()
			go a.gracefulShutdown()
			a.execution.tolerantMode = true
		case sig := <-syscallC:
			runDone = nil
			signal.Stop(syscallC)

			a.logger.Info(`graceful shutdown started by syscall`,
				"application", a.name,
				`syscall`, sig.String())

			a.execution.initRunCancel()
			go a.gracefulShutdown()
			a.execution.tolerantMode = true
		case <-a.shutdown.shutdownDone:
			break CycleLable
		case <-execDone:
			execDone = nil
			a.logger.Debug(`execution finished`,
				"application", a.name)
		}
	}
}
