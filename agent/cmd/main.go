package main

import (
	"agent/internal/client"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"os"
	"pkg/api"
	"pkg/logger"
)

func main() {
	clientLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := logger.WithLogger(context.Background(), clientLogger)
	log := logger.GetLogger(ctx)
	conn, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to orch_grpc server: ", err)
		return
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Error("Failed to close connection to orch_grpc server: ", err)
		}
	}(conn)

	orchClient := api.NewOrchestratorClient(conn)
	log.Info("Starting orchestrator client")

	agent := client.NewAgentClient(orchClient)

	client.ManageTasks(ctx, agent)
}
