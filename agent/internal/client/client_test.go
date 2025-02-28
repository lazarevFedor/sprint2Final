package client

import (
	"agent/internal/entities"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	logger2 "pkg/logger"
	"testing"
	"time"
)

func TestManageTasks(t *testing.T) {
	tests := []struct {
		name           string
		computingPower string
		serverResponse entities.AgentResponse
		serverStatus   int
		wantTaskSent   bool
	}{
		{
			name:           "Server returns 404",
			computingPower: "1",
			serverResponse: entities.AgentResponse{},
			serverStatus:   http.StatusNotFound,
			wantTaskSent:   false,
		},
		{
			name:           "Server returns 500",
			computingPower: "1",
			serverResponse: entities.AgentResponse{},
			serverStatus:   http.StatusInternalServerError,
			wantTaskSent:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					w.WriteHeader(tt.serverStatus)
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			defer func(oldURL string) { os.Setenv("ORCHESTRATOR_URL", oldURL) }(os.Getenv("ORCHESTRATOR_URL"))
			os.Setenv("ORCHESTRATOR_URL", server.URL)

			defer func(oldPower string) { os.Setenv("COMPUTING_POWER", oldPower) }(os.Getenv("COMPUTING_POWER"))
			os.Setenv("COMPUTING_POWER", tt.computingPower)

			var logBuf bytes.Buffer
			clientLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			ctx := logger2.WithLogger(context.Background(), clientLogger)

			taskChan := make(chan entities.AgentResponse, 1)
			go func() {
				ManageTasks(ctx)
			}()

			time.Sleep(6 * time.Second)

			select {
			case task := <-taskChan:
				if !tt.wantTaskSent {
					t.Errorf("Task was sent to channel, but it shouldn't have been: %+v", task)
				}
			default:
				if tt.wantTaskSent {
					t.Errorf("No task was sent to channel, but it should have been")
				}
			}
		})
	}
}

func TestSolveTask(t *testing.T) {
	tests := []struct {
		name           string
		task           entities.AgentResponse
		wantResult     float64
		wantError      error
		wantStatusCode int
	}{
		{
			name:           "Successful addition",
			task:           entities.AgentResponse{Id: 1, Arg1: 5, Arg2: 3, Operation: "+", OperationTime: 10},
			wantResult:     8.0,
			wantError:      nil,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Successful multiplication",
			task:           entities.AgentResponse{Id: 2, Arg1: 4, Arg2: 3, Operation: "*", OperationTime: 20},
			wantResult:     12.0,
			wantError:      nil,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Wrong operator",
			task:           entities.AgentResponse{Id: 3, Arg1: 5, Arg2: 2, Operation: "%", OperationTime: 10},
			wantResult:     0.0,
			wantError:      errors.New("wrong operator"),
			wantStatusCode: 0,
		},
		{
			name:           "Server error",
			task:           entities.AgentResponse{Id: 4, Arg1: 2, Arg2: 2, Operation: "-", OperationTime: 5},
			wantResult:     0.0,
			wantError:      nil,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "POST" {
					w.WriteHeader(tt.wantStatusCode)
					if tt.wantStatusCode == http.StatusOK {
						var req entities.AgentRequest
						json.NewDecoder(r.Body).Decode(&req)
						if req.Result != tt.wantResult {
							t.Errorf("Sent result = %v, want %v", req.Result, tt.wantResult)
						}
					}
				}
			}))
			defer server.Close()

			var logBuf bytes.Buffer
			clientLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			ctx := logger2.WithLogger(context.Background(), clientLogger)

			client := &http.Client{}
			solveTask(client, tt.task, ctx)
		})
	}
}

func TestWorker(t *testing.T) {
	t.Run("Worker processes task", func(t *testing.T) {
		taskChan := make(chan entities.AgentResponse, 1)
		done := make(chan struct{})

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		var logBuf bytes.Buffer
		clientLogger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), clientLogger)

		go func() {
			worker(&http.Client{}, taskChan, ctx)
			close(done)
		}()

		task := entities.AgentResponse{Id: 1, Arg1: 2, Arg2: 3, Operation: "+", OperationTime: 10}
		taskChan <- task
		close(taskChan)

		<-done
	})
}
