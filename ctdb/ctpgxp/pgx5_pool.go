package ctpgxp

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/surkovvs/ct/ctifaces"
	"github.com/surkovvs/ct/internal/tools"
)

type Pool struct {
	Cfg  *pgxpool.Config
	pgxp *pgxpool.Pool
}

func New(cfg ctifaces.SQLConfigurator) (*Pool, error) {
	pgxpCfg, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("pool parse config: %w", err)
	}

	if cfg.LogQueriesEnabled() {
		var logger ctifaces.Logger
		if cLogger := cfg.GetLogger(); cLogger != nil {
			logger = cLogger
		} else {
			logger = tools.NewDefaultLogger()
		}
		pgxpCfg.ConnConfig.Tracer = tracer{logger}
	}

	cfgPool := cfg.GetPGPoolConfig()
	tools.SetIfNotNil(&pgxpCfg.MaxConnLifetime, cfgPool.GetMaxConnLifetime())
	tools.SetIfNotNil(&pgxpCfg.MaxConnLifetimeJitter, cfgPool.GetMaxConnLifetimeJitter())
	tools.SetIfNotNil(&pgxpCfg.MaxConnIdleTime, cfgPool.GetMaxConnIdleTime())
	tools.SetIfNotNil(&pgxpCfg.MaxConns, cfgPool.GetMaxConns())
	tools.SetIfNotNil(&pgxpCfg.MinConns, cfgPool.GetMinConns())
	tools.SetIfNotNil(&pgxpCfg.HealthCheckPeriod, cfgPool.GetHealthCheckPeriod())

	return &Pool{
		Cfg: pgxpCfg,
	}, nil
}

func (pool *Pool) GetPool() *pgxpool.Pool {
	return pool.pgxp
}

func (pool *Pool) Init(ctx context.Context) error {
	var err error
	pool.pgxp, err = pgxpool.NewWithConfig(ctx, pool.Cfg)
	if err != nil {
		return fmt.Errorf("new pgx pool: %w", err)
	}
	err = pool.pgxp.Ping(ctx)
	if err != nil {
		return fmt.Errorf("pgx pool ping: %w", err)
	}
	return nil
}

func (pool *Pool) Shutdown(_ context.Context) error {
	pool.pgxp.Close()
	return nil
}

func (pool *Pool) GetModuleNamePrefix() string {
	return "sql_pgxp"
}

func (pool *Pool) PreidentifyModuleGroup() string {
	return "egress"
}
