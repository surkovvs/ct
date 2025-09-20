package tools

import (
	"log/slog"
	"os"
)

func SetIfNotNil[T any](a *T, b *T) {
	if b != nil {
		*a = *b
	}
}

func NewDefaultLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}))
}
