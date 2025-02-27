package main

import (
	"agent/internal/client"
	"context"
	"log/slog"
	"os"
	"pkg/logger"
)

func main() {
	clientLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := logger.WithLogger(context.Background(), clientLogger)
	client.ManageTasks(ctx)
}
