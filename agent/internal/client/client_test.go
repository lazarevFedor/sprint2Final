package client

import (
	"agent/internal/entities"
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"os"
	"pkg/api"
	logger2 "pkg/logger"
	"testing"
	"time"
)

// mockOrchestratorClient struct for testing client.go
type mockOrchestratorClient struct {
	getTaskFunc    func(ctx context.Context, in *api.GetTaskRequest, opts ...grpc.CallOption) (*api.GetTaskResponse, error)
	postTaskCalled bool
	postTaskID     int32
	postTaskResult float32
	postTaskError  error
}

// GetTask imitates server handler
func (m *mockOrchestratorClient) GetTask(ctx context.Context, in *api.GetTaskRequest, opts ...grpc.CallOption) (*api.GetTaskResponse, error) {
	if m.getTaskFunc != nil {
		return m.getTaskFunc(ctx, in, opts...)
	}
	return nil, nil
}

// PostTask imitates server handler
func (m *mockOrchestratorClient) PostTask(ctx context.Context, in *api.PostTaskRequest, opts ...grpc.CallOption) (*api.PostTaskResponse, error) {
	m.postTaskCalled = true
	m.postTaskID = in.Id
	m.postTaskResult = in.Result
	return &api.PostTaskResponse{}, m.postTaskError
}

// TestSolveTask_Success tests the happy path where calculation and posting succeed
func TestSolveTask_Success(t *testing.T) {
	mockClient := &mockOrchestratorClient{postTaskError: nil}
	agent := NewAgentClient(mockClient)
	ctx := logger2.WithLogger(context.Background(), slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	task := entities.AgentResponse{
		Id:            1,
		Arg1:          2.0,
		Arg2:          3.0,
		Operation:     "+",
		OperationTime: 100,
	}

	solveTask(agent, task, ctx)

	assert.True(t, mockClient.postTaskCalled)
	assert.Equal(t, int32(1), mockClient.postTaskID)
	assert.Equal(t, float32(5.0), mockClient.postTaskResult)
}

// TestSolveTask_PostTaskFails tests when PostTask fails
func TestSolveTask_PostTaskFails(t *testing.T) {
	mockClient := &mockOrchestratorClient{postTaskError: errors.New("post task error")}
	agent := NewAgentClient(mockClient)
	ctx := logger2.WithLogger(context.Background(), slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	task := entities.AgentResponse{
		Id:            1,
		Arg1:          2.0,
		Arg2:          3.0,
		Operation:     "+",
		OperationTime: 100,
	}

	solveTask(agent, task, ctx)

	assert.True(t, mockClient.postTaskCalled)
}

// TestWorker tests that the worker processes a task from the channel
func TestWorker(t *testing.T) {

	mockClient := &mockOrchestratorClient{postTaskError: nil}
	agent := NewAgentClient(mockClient)
	taskChan := make(chan entities.AgentResponse, 1)
	ctx := logger2.WithLogger(context.Background(), slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	go worker(agent, taskChan, ctx)

	task := entities.AgentResponse{
		Id:            1,
		Arg1:          2.0,
		Arg2:          3.0,
		Operation:     "+",
		OperationTime: 100,
	}
	taskChan <- task

	time.Sleep(100 * time.Millisecond) // Give worker time to process
	assert.True(t, mockClient.postTaskCalled)
	assert.Equal(t, int32(1), mockClient.postTaskID)
	assert.Equal(t, float32(5.0), mockClient.postTaskResult)
}

// TestManageTasks tests that ManageTasks starts workers and processes a task
func TestManageTasks(t *testing.T) {
	os.Setenv("COMPUTING_POWER", "1")

	taskReturned := false
	mockClient := &mockOrchestratorClient{
		getTaskFunc: func(ctx context.Context, in *api.GetTaskRequest, opts ...grpc.CallOption) (*api.GetTaskResponse, error) {
			if !taskReturned {
				taskReturned = true
				return &api.GetTaskResponse{Id: 1, Arg1: 2.0, Arg2: 3.0, Operation: "+", OperationTime: 100}, nil
			}
			return nil, status.Error(codes.NotFound, "no tasks")
		},
		postTaskError: nil,
	}

	agent := NewAgentClient(mockClient)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = logger2.WithLogger(ctx, slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	go ManageTasks(ctx, agent)

	timeout := time.After(10 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for PostTask")
		default:
			if mockClient.postTaskCalled {
				assert.Equal(t, int32(1), mockClient.postTaskID)
				assert.Equal(t, float32(5.0), mockClient.postTaskResult)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
