package main

import (
	"context"
	"log/slog"
	"orchestrator/internal/server"
	"os"
	"pkg/logger"
)

func main() {
	serverLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := logger.WithLogger(context.Background(), serverLogger)
	err := server.StartServer(ctx)
	if err != nil {
		return
	}
}
