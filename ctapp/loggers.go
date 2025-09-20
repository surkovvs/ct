package ctapp

import (
	"log/slog"
	"os"

	"github.com/surkovvs/ct/ctifaces"
)

type logStub struct{}

func (logStub) Debug(string, ...any) {}

func (logStub) Info(string, ...any) {}

func (logStub) Warn(string, ...any) {}

func (logStub) Error(string, ...any) {}

type logWrap struct {
	logger ctifaces.Logger
}

func defaultLog() logWrap {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return logWrap{
		logger: logger,
	}
}

func (l logWrap) Debug(msg string, args ...any) {
	l.logger.Debug(`[ct-app] `+msg, args...)
}

func (l logWrap) Info(msg string, args ...any) {
	l.logger.Info(`[ct-app] `+msg, args...)
}

func (l logWrap) Warn(msg string, args ...any) {
	l.logger.Warn(`[ct-app] `+msg, args...)
}

func (l logWrap) Error(msg string, args ...any) {
	l.logger.Error(`[ct-app] `+msg, args...)
}
