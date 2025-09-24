package ctapp

import (
	"os"
	"time"

	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.AppConfigurator = (*Config)(nil)

type Config struct {
	Silient         bool
	TolerantMode    bool
	Name            *string
	InitTimeout     *time.Duration
	ShutdownTimeout *time.Duration
}

func (c Config) IsAppSilientMode() bool {
	return c.Silient
}

func (c Config) IsAppTolerantMode() bool {
	return c.TolerantMode
}

func (c Config) GetApplicationName() *string {
	return c.Name
}

func (c Config) GetAppInitTimeout() *time.Duration {
	return c.InitTimeout
}

func (c Config) GetAppShutdownTimeout() *time.Duration {
	return c.ShutdownTimeout
}

type AppOption func(*App)

func WithLogger(logger ctifaces.Logger) AppOption {
	return func(a *App) {
		a.logger = logger
	}
}

func WithProvidedSigs(sigs ...os.Signal) AppOption {
	return func(a *App) {
		a.shutdown.sigs = sigs
	}
}

func WithConfig(cfg ctifaces.AppConfigurator) AppOption {
	return func(a *App) {
		if cfg.IsAppSilientMode() {
			a.logger = logStub{}
		}
		if cfg.IsAppTolerantMode() {
			a.execution.tolerantMode = true
		}
		if cfg.GetApplicationName() != nil {
			a.name = *cfg.GetApplicationName()
		}
		if cfg.GetAppInitTimeout() != nil {
			a.execution.initTimeout = cfg.GetAppInitTimeout()
		}
		if cfg.GetAppShutdownTimeout() != nil {
			a.shutdown.timeout = cfg.GetAppShutdownTimeout()
		}
	}
}

func (a *App) defaultSettingsCheckAndApply() {
	if a.name == "" {
		a.name = `unnamed`
	}

	if a.logger == nil {
		a.logger = defaultLog()
	}

	if a.shutdown.sigs == nil {
		a.shutdown.sigs = DefaultProvidedSigs
	}
	if a.shutdown.timeout == nil {
		a.shutdown.timeout = &DefaultShutdownTimeout
	}
}
