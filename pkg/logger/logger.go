package logger

import (
	"context"
	"log/slog"
)

type key string

const loggerKey = key("logger")

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger := ctx.Value(loggerKey)
	return logger.(*slog.Logger)
}
