package ctifaces

import (
	"context"
	"time"
)

type (
	Healthchecker interface {
		Healthcheck(ctx context.Context) error
	}
	Initializer interface {
		Init(ctx context.Context) error
	}
	Runner interface {
		Run(ctx context.Context) error
	}
	Shutdowner interface {
		Shutdown(ctx context.Context) error
	}
)

type AppConfigurator interface {
	IsAppSilientMode() bool
	IsAppTolerantMode() bool
	GetApplicationName() *string
	GetAppInitTimeout() *time.Duration
	GetAppShutdownTimeout() *time.Duration
}

type NamePrefixer interface {
	GetModuleNamePrefix() string
}

type GroupIdetifer interface {
	PreidentifyModuleGroup() string
}
