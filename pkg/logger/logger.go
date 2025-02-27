package logger

import (
	"context"
	"log/slog"
)

type key string

const loggerKey = key("logger")

func withLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func getLogger(ctx context.Context) *slog.Logger {
	logger := ctx.Value(loggerKey)
	return logger.(*slog.Logger)
}
