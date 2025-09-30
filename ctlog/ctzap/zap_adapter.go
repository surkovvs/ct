package ctzap

import (
	"log/slog"
	"os"

	"github.com/surkovvs/ct/ctifaces"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_ ctifaces.Logger         = (*ZapAdapter)(nil)
	_ ctifaces.NameableLogger = (*ZapAdapter)(nil)
)

type ZapAdapter struct {
	sl *zap.SugaredLogger
}

func New(cfg ctifaces.LogConfigurator) ZapAdapter {
	var encoder zapcore.Encoder
	if cfg.IsLogDevMode() {
		zapCfg := zap.NewDevelopmentEncoderConfig()
		if cfg.IsLogColored() {
			zapCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}
		encoder = zapcore.NewConsoleEncoder(zapCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	}

	return ZapAdapter{
		sl: zap.New(
			zapcore.NewCore(encoder, os.Stdout, toZapLVL(cfg.GetLogLvl())),
		).
			Sugar(),
	}
}

func NewZapBased(logger *zap.SugaredLogger) ZapAdapter {
	return ZapAdapter{
		sl: logger,
	}
}

func toZapLVL(l slog.Level) zapcore.Level {
	switch l {
	case slog.LevelDebug:
		return zapcore.DebugLevel
	case slog.LevelInfo:
		return zapcore.InfoLevel
	case slog.LevelWarn:
		return zapcore.WarnLevel
	case slog.LevelError:
		return zapcore.ErrorLevel
	}
	return zapcore.InvalidLevel
}

func (za ZapAdapter) Debug(msg string, args ...any) {
	za.sl.Debugw(msg, args...)
}

func (za ZapAdapter) Info(msg string, args ...any) {
	za.sl.Infow(msg, args...)
}

func (za ZapAdapter) Warn(msg string, args ...any) {
	za.sl.Warnw(msg, args...)
}

func (za ZapAdapter) Error(msg string, args ...any) {
	za.sl.Errorw(msg, args...)
}

func (za ZapAdapter) Named(name string) ctifaces.NameableLogger {
	return ZapAdapter{
		sl: za.sl.Named(name),
	}
}

func (za ZapAdapter) GetZap() *zap.SugaredLogger {
	return za.sl
}
