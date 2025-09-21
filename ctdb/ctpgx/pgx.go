package ctpgx

import (
	"context"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/surkovvs/ct/ctifaces"
	"github.com/surkovvs/ct/internal/tools"
)

type Pool struct {
	Cfg  pgx.ConnConfig
	pool *pgx.ConnPool
}

func New(cfg ctifaces.SQLConfigurator) (*Pool, error) {
	connCfg, err := pgx.ParseConnectionString(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("pgx parse dsn: %w", err)
	}

	connCfg.TLSConfig = cfg.GetTLS()

	if cfg.LogQueriesEnabled() {
		var logger ctifaces.Logger
		if cLogger := cfg.GetLogger(); cLogger != nil {
			logger = cLogger
		} else {
			logger = tools.NewDefaultLogger()
		}
		connCfg.LogLevel = pgx.LogLevelTrace
		connCfg.Logger = logAdapter{logger}
	}

	return &Pool{
		Cfg: connCfg,
	}, nil
}

func (pool *Pool) GetPool() *pgx.ConnPool {
	return pool.pool
}

func (pool *Pool) Init(ctx context.Context) error {
	var err error
	pool.pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pool.Cfg,
	})
	if err != nil {
		return fmt.Errorf("pgx new conn pool: %w", err)
	}

	conn, err := pool.pool.Acquire()
	if err != nil {
		return fmt.Errorf("pgx acquire: %w", err)
	}
	defer pool.pool.Release(conn)
	err = conn.Ping(ctx)
	if err != nil {
		return fmt.Errorf("pgx ping: %w", err)
	}

	return nil
}

func (pool *Pool) Shutdown(_ context.Context) error {
	pool.pool.Close()
	return nil
}

func (pool *Pool) GetModuleNamePrefix() string {
	return "sql_pgx"
}

func (pool *Pool) PreidentifyModuleGroup() string {
	return "background_sync"
}
