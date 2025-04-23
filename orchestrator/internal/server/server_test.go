package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"orchestrator/internal/entities"
	"pkg"
	logger2 "pkg/logger"
	"sync"
	"testing"
)

type mockParsersTree struct {
	data map[int]*chan float64
	sync.Mutex
}

func (m *mockParsersTree) Insert(id int, ch *chan float64) {
	m.Lock()
	defer m.Unlock()
	m.data[id] = ch
}

func (m *mockParsersTree) Delete(id int) error {
	m.Lock()
	defer m.Unlock()
	delete(m.data, id)
	return nil
}

func TestIsValidExpression(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       bool
	}{
		{name: "Valid expression", expression: "2 + 3", want: true},
		{name: "Valid with parentheses", expression: "(2 + 3) * 4", want: true},
		{name: "Invalid with letters", expression: "2 + a", want: false},
		{name: "Empty string", expression: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidExpression(tt.expression); got != tt.want {
				t.Errorf("isValidExpression(%q) = %v, want %v", tt.expression, got, tt.want)
			}
		})
	}
}

func TestAgentMiddleware(t *testing.T) {
	getHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("GET"))
		if err != nil {
			return
		}
	})
	postHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("POST"))
		if err != nil {
			return
		}
	})

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantBody   string
	}{
		{name: "GET request", method: "GET", wantStatus: http.StatusOK, wantBody: "GET"},
		{name: "POST request", method: "POST", wantStatus: http.StatusOK, wantBody: "POST"},
		{name: "PUT request", method: "PUT", wantStatus: http.StatusMethodNotAllowed, wantBody: "Method not allowed\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/internal/task", nil)
			w := httptest.NewRecorder()
			handler := agentMiddleware(getHandler, postHandler)
			handler.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("agentMiddleware status = %v, want %v", w.Code, tt.wantStatus)
			}
			if w.Body.String() != tt.wantBody {
				t.Errorf("agentMiddleware body = %q, want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestGetHandler(t *testing.T) {
	t.Run("Task available", func(t *testing.T) {
		// Устанавливаем очередь с задачей
		queue := &pkg.Queue{}
		task := entities.Task{Id: 1, Arg1: 2, Arg2: 3, Operation: "+"}
		queue.Enqueue(task)
		entities.Tasks = queue

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), logger)

		req := httptest.NewRequest("GET", "/internal/task", nil)
		w := httptest.NewRecorder()
		getHandler(ctx).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("getHandler status = %v, want %v", w.Code, http.StatusOK)
		}
		var receivedTask entities.Task
		if err := json.Unmarshal(w.Body.Bytes(), &receivedTask); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}
		if receivedTask.Id != 1 {
			t.Errorf("getHandler task.Id = %v, want 1", receivedTask.Id)
		}
	})

	t.Run("No tasks", func(t *testing.T) {
		// Устанавливаем пустую очередь
		entities.Tasks = &pkg.Queue{}

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), logger)

		req := httptest.NewRequest("GET", "/internal/task", nil)
		w := httptest.NewRecorder()
		getHandler(ctx).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("getHandler status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}
