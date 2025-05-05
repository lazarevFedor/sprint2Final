package grpc_server

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	obj "orchestrator/internal/entities"
	"os"
	"pkg/api"
	"pkg/logger"
)

type Server struct {
	api.OrchestratorServer
}

func New() *Server {
	return &Server{}
}

func (s *Server) GetTask(_ context.Context, _ *api.GetTaskRequest) (*api.GetTaskResponse, error) {
	serverLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := logger.WithLogger(context.Background(), serverLogger)
	log := logger.GetLogger(ctx)

	if obj.Tasks.IsEmpty() {
		return nil, status.Error(codes.NotFound, "No available tasks")
	}
	task := obj.Tasks.Dequeue().(obj.Task)
	log.Info("Task dequeued with Id", task.Id)
	return &api.GetTaskResponse{
		Id:            int32(task.Id),
		Arg1:          float32(task.Arg1),
		Arg2:          float32(task.Arg2),
		Operation:     task.Operation,
		OperationTime: int32(task.OperationTime),
	}, nil
}

func (s *Server) PostTask(_ context.Context, request *api.PostTaskRequest) (*api.PostTaskResponse, error) {
	serverLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := logger.WithLogger(context.Background(), serverLogger)
	log := logger.GetLogger(ctx)
	node := obj.ParsersTree.Search(int(request.Id))
	if node == nil {
		log.Error("Node not found")
		return nil, status.Error(codes.NotFound, "Task not found")
	}
	ch := node.Value.(*chan float64)
	*ch <- float64(request.Result)
	log.Info("PostTask dequeued with Id", request.Id)
	return &api.PostTaskResponse{}, nil
}
