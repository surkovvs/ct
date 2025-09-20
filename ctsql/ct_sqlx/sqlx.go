package ctsqlx

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/surkovvs/ct/ctifaces"
	"github.com/surkovvs/ct/internal/tools"
)

type Database struct {
	*sqlx.DB
}

func New(cfg ctifaces.SQLConfigurator) (*Database, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	if cfg.LogQueriesEnabled() {
		var logger ctifaces.Logger
		if cLogger := cfg.GetLogger(); cLogger != nil {
			logger = cLogger
		} else {
			logger = tools.NewDefaultLogger()
		}
		db = sqldblogger.OpenDriver(cfg.GetDSN(), db.Driver(), logAdapter{logger})
	}

	return &Database{sqlx.NewDb(db, "postgres")}, nil
}

func (db *Database) Init(ctx context.Context) error {
	return db.PingContext(ctx)
}

func (db *Database) Shutdown(_ context.Context) error {
	return db.Close()
}

func (db *Database) GetModuleNamePrefix() string {
	return "sqlx"
}

func (db *Database) PreidentifyModuleGroup() string {
	return "background_sync"
}

type logAdapter struct {
	ctifaces.Logger
}

func (l logAdapter) Log(_ context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	fields := make([]any, 0, len(data))
	for k, v := range data {
		fields = append(fields, k, v)
	}
	switch level {
	case sqldblogger.LevelTrace:
		l.Debug(msg, fields...)
	case sqldblogger.LevelDebug:
		l.Debug(msg, fields...)
	case sqldblogger.LevelInfo:
		l.Info(msg, fields...)
	case sqldblogger.LevelError:
		l.Error(msg, fields...)
	}
}

// LogLevelTrace = 6
// LogLevelDebug = 5
// LogLevelInfo  = 4
// LogLevelWarn  = 3
// LogLevelError = 2
// LogLevelNone  = 1
