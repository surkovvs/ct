package ctzerolog

import (
	"github.com/rs/zerolog"
	"github.com/surkovvs/ct/ctifaces"
)

var _ ctifaces.Logger = (*ZerologAdapter)(nil)

type ZerologAdapter struct {
	*zerolog.Logger
}

// TODO: refactor as zap adapter implemented

func NewZerologAdapter(logger *zerolog.Logger) ZerologAdapter {
	return ZerologAdapter{logger}
}

func (zla ZerologAdapter) Debug(msg string, args ...any) {
	zla.Logger.Debug().Fields(args).Msg(msg)
}

func (zla ZerologAdapter) Info(msg string, args ...any) {
	zla.Logger.Info().Fields(args).Msg(msg)
}

func (zla ZerologAdapter) Warn(msg string, args ...any) {
	zla.Logger.Warn().Fields(args).Msg(msg)
}

func (zla ZerologAdapter) Error(msg string, args ...any) {
	zla.Logger.Error().Fields(args).Msg(msg)
}
