package ctlog

import (
	"log/slog"
	"strings"
)

type ConfigLog struct {
	Level   any
	Develop bool // if not - json encoder will be used
	Colored bool
}

type Level int

const (
	LevelDebugStr string = "debug"
	LevelInfoStr  string = "info"
	LevelWarnStr  string = "warn"
	LevelErrorStr string = "error"
)

func (c ConfigLog) GetLogLvl() slog.Level {
	switch toParse := c.Level.(type) {
	case slog.Level:
		return slog.LevelDebug // TODO: try it out
	case string:
		switch {
		case strings.EqualFold(LevelDebugStr, toParse):
			return slog.LevelDebug
		case strings.EqualFold(LevelInfoStr, toParse):
			return slog.LevelInfo
		case strings.EqualFold(LevelWarnStr, toParse):
			return slog.LevelWarn
		case strings.EqualFold(LevelErrorStr, toParse):
			return slog.LevelError
		}
	}
	return 0
}

func (c ConfigLog) IsLogDevMode() bool {
	return c.Develop
}

func (c ConfigLog) IsLogColored() bool {
	return c.Colored
}
