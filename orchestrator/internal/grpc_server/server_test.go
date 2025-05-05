package grpc_server

import (
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"orchestrator/internal/entities"
	"pkg/api"
	"testing"
)

func TestGetTask_EmptyQueue(t *testing.T) {
	server := New()

	resp, err := server.GetTask(context.Background(), &api.GetTaskRequest{})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestGetTask_NonEmptyQueue(t *testing.T) {
	server := New()
	task := entities.Task{Id: 1, Arg1: 2.0, Arg2: 3.0, Operation: "+", OperationTime: 100}
	entities.Tasks.Enqueue(task)

	resp, err := server.GetTask(context.Background(), &api.GetTaskRequest{})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(1), resp.Id)
	assert.Equal(t, float32(2.0), resp.Arg1)
	assert.Equal(t, float32(3.0), resp.Arg2)
	assert.Equal(t, "+", resp.Operation)
	assert.Equal(t, int32(100), resp.OperationTime)
}

func TestPostTask_TaskNotFound(t *testing.T) {
	server := New()

	resp, err := server.PostTask(context.Background(), &api.PostTaskRequest{Id: 1, Result: 5.0})

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestPostTask_TaskFound(t *testing.T) {
	server := New()
	ch := make(chan float64, 1)
	entities.ParsersTree.Insert(1, &ch)

	go func() {
		resp, err := server.PostTask(context.Background(), &api.PostTaskRequest{Id: 1, Result: 5.0})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	}()

	result := <-ch
	assert.Equal(t, 5.0, result)
}
