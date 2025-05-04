package client

import (
	"agent/internal/demon"
	"agent/internal/entities"
	"context"
	"os"
	"pkg/api"
	logger2 "pkg/logger"
	"strconv"
	"time"
)

type AgentClient struct {
	client api.OrchestratorClient
}

func NewAgentClient(client api.OrchestratorClient) *AgentClient {
	return &AgentClient{client: client}
}

// ManageTasks is a function that manages tasks
func ManageTasks(ctx context.Context, agent *AgentClient) {
	logger := logger2.GetLogger(ctx)
	logger.Info("ManageTasks: Start")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	computingPower, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil || computingPower <= 0 {
		computingPower = 1
	}
	taskChan := make(chan entities.AgentResponse, 1)

	for i := 0; i < computingPower; i++ {
		go worker(agent, taskChan, ctx)
	}

	for range ticker.C {
		taskAccepted, err := agent.client.GetTask(ctx, &api.GetTaskRequest{})
		if err != nil {
			continue
		}

		task := entities.AgentResponse{
			Id:            int(taskAccepted.Id),
			Arg1:          float64(taskAccepted.Arg1),
			Arg2:          float64(taskAccepted.Arg2),
			Operation:     taskAccepted.Operation,
			OperationTime: int(taskAccepted.OperationTime),
		}

		logger.Info("ManageTasks: Task accepted:", task.Id)

		taskChan <- task
		logger.Info("ManageTasks: Task received")
	}
}

// worker is a function that processes tasks
func worker(agent *AgentClient, taskChan <-chan entities.AgentResponse, ctx context.Context) {
	for task := range taskChan {
		solveTask(agent, task, ctx)
	}
}

// solveTask is a function that solves the task
func solveTask(agent *AgentClient, task entities.AgentResponse, ctx context.Context) {
	logger := logger2.GetLogger(ctx)
	result, err := demon.CalculateExpression(task.Arg1, task.Arg2, task.Operation, task.OperationTime)
	if err != nil {
		logger.Error("solveTask: calculating expression error:", "err", err)
		return
	}

	_, err = agent.client.PostTask(ctx, &api.PostTaskRequest{
		Id:     int32(task.Id),
		Result: float32(result),
	})
	if err != nil {
		logger.Error("solveTask: post task error:", "err", err)
		return
	}
	logger.Info("solveTask: Task solved")
	logger.Info("Id:", task.Id, "Result:", result)
}
