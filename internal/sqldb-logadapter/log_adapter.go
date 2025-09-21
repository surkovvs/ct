package sqldblogadapter

import (
	"context"

	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/surkovvs/ct/ctifaces"
)

type LogAdapter struct {
	ctifaces.Logger
}

func (l LogAdapter) Log(_ context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
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
