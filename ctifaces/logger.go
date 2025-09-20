package ctifaces

import "log/slog"

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

type LogConfigurator interface {
	GetLogLvl() slog.Level
	IsLogDevMode() bool
	IsLogColored() bool
}
