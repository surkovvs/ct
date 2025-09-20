package ctpgx

import (
	"github.com/jackc/pgx"
	"github.com/surkovvs/ct/ctifaces"
)

// TODO: implement methods
func New(cfg ctifaces.SQLConfigurator) (*pgx.ConnPool, error) {
	connCfg, err := pgx.ParseConnectionString(cfg.GetDSN())
	if err != nil {
		return nil, err
	}

	logger := cfg.GetLogger()
	if logger != nil {
		connCfg.LogLevel = 6
		connCfg.Logger = logAdapter{logger}
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: connCfg,
	})
	if err != nil {
		return nil, err
	}
	return pool, nil
}

type logAdapter struct {
	ctifaces.Logger
}

func (l logAdapter) Log(level pgx.LogLevel, msg string, data map[string]any) {
	fields := make([]any, 0, len(data))
	for k, v := range data {
		fields = append(fields, k, v)
	}
	switch level {
	case pgx.LogLevelTrace:
		l.Debug(msg, fields...)
	case pgx.LogLevelDebug:
		l.Debug(msg, fields...)
	case pgx.LogLevelInfo:
		l.Info(msg, fields...)
	case pgx.LogLevelWarn:
		l.Warn(msg, fields...)
	case pgx.LogLevelError:
		l.Error(msg, fields...)
	}
}
