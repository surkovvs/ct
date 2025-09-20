package ctifaces

import (
	"crypto/tls"
	"time"
)

type SQLConfigurator interface {
	GetDSN() string
	GetLogger() Logger
	LogQueriesEnabled() bool
	GetPGPoolConfig() PGPoolConfigurator
	GetTLS() *tls.Config
}

type PGPoolConfigurator interface {
	GetMaxConnLifetime() *time.Duration
	GetMaxConnLifetimeJitter() *time.Duration
	GetMaxConnIdleTime() *time.Duration
	GetMaxConns() *int32
	GetMinConns() *int32
	GetHealthCheckPeriod() *time.Duration
}
