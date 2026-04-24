package ctsqlx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // driver
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/surkovvs/ct/ctifaces"
	sqldblogadapter "github.com/surkovvs/ct/internal/sqldb-logadapter"
	"github.com/surkovvs/ct/internal/tools"
)

type Database struct {
	*sqlx.DB
}

func New(cfg ctifaces.SQLConfigurator) (*Database, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("sql open: %w", err)
	}

	if cfg.LogQueriesEnabled() {
		var logger ctifaces.Logger
		if cLogger := cfg.GetLogger(); cLogger != nil {
			logger = cLogger
		} else {
			logger = tools.NewDefaultLogger()
		}
		db = sqldblogger.OpenDriver(cfg.GetDSN(), db.Driver(),
			sqldblogadapter.LogAdapter{Logger: logger})
	}

	return &Database{sqlx.NewDb(db, "postgres")}, nil
}

func (db *Database) Init(ctx context.Context) error {
	err := db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("sqlx ping: %w", err)
	}
	return nil
}

func (db *Database) Shutdown(_ context.Context) error {
	err := db.Close()
	if err != nil {
		return fmt.Errorf("sqlx close: %w", err)
	}
	return nil
}

func (db *Database) GetModuleNamePrefix() string {
	return "sqlx"
}

func (db *Database) PreidentifyModuleGroup() string {
	return "egress"
}
