package ctdb

import (
	"time"

	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.PGPoolConfigurator = PGPoolConfig{}

type PGPoolConfig struct {
	LogConnectOperations  bool
	MaxConnLifetime       *time.Duration
	MaxConnLifetimeJitter *time.Duration
	MaxConnIdleTime       *time.Duration
	MaxConns              *int32
	MinConns              *int32
	HealthCheckPeriod     *time.Duration
}

func (c PGPoolConfig) GetMaxConnLifetime() *time.Duration {
	return c.MaxConnLifetime
}

func (c PGPoolConfig) GetMaxConnLifetimeJitter() *time.Duration {
	return c.MaxConnLifetimeJitter
}

func (c PGPoolConfig) GetMaxConnIdleTime() *time.Duration {
	return c.MaxConnIdleTime
}

func (c PGPoolConfig) GetMaxConns() *int32 {
	return c.MaxConns
}

func (c PGPoolConfig) GetMinConns() *int32 {
	return c.MinConns
}

func (c PGPoolConfig) GetHealthCheckPeriod() *time.Duration {
	return c.HealthCheckPeriod
}
