package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"orchestrator/internal/entities"
	pkg "pkg" // Импорт pkg для SafeMap
	logger2 "pkg/logger"
	"strconv"
	"sync"
	"testing"
)

// Моки для orchestratorState (кроме Expressions)
type mockTasks struct {
	queue []entities.Task
	sync.Mutex
}

func (m *mockTasks) Enqueue(t entities.Task) {
	m.Lock()
	defer m.Unlock()
	m.queue = append(m.queue, t)
}

func (m *mockTasks) Dequeue() entities.Task {
	m.Lock()
	defer m.Unlock()
	if len(m.queue) == 0 {
		return entities.Task{}
	}
	t := m.queue[0]
	m.queue = m.queue[1:]
	return t
}

func (m *mockTasks) IsEmpty() bool {
	m.Lock()
	defer m.Unlock()
	return len(m.queue) == 0
}

type mockTasksLastID struct {
	value int
	sync.Mutex
}

func (m *mockTasksLastID) GetValue() int {
	m.Lock()
	defer m.Unlock()
	return m.value
}

func (m *mockTasksLastID) Increment() {
	m.Lock()
	defer m.Unlock()
	m.value++
}

type mockParsersTree struct {
	data map[int]*chan float64
	sync.Mutex
}

func (m *mockParsersTree) Insert(id int, ch *chan float64) {
	m.Lock()
	defer m.Unlock()
	m.data[id] = ch
}

//func (m *mockParsersTree) Search(id int) *entities.ParserNode {
//	m.Lock()
//	defer m.Unlock()
//	if ch, ok := m.data[id]; ok {
//		return &entities.ParserNode{Value: ch}
//	}
//	return nil
//}

func (m *mockParsersTree) Delete(id int) error {
	m.Lock()
	defer m.Unlock()
	delete(m.data, id)
	return nil
}

// Глобальные моки с переименованной структурой
var (
	mockWg            sync.WaitGroup
	orchestratorState = struct { // Переименована с obj на orchestratorState
		Expressions *pkg.SafeMap
		Tasks       *mockTasks
		TasksLastID *mockTasksLastID
		ParsersTree *mockParsersTree
		Wg          *sync.WaitGroup
	}{
		Expressions: pkg.NewSafeMap(),
		Tasks:       &mockTasks{queue: []entities.Task{}},
		TasksLastID: &mockTasksLastID{value: 1},
		ParsersTree: &mockParsersTree{data: make(map[int]*chan float64)},
		Wg:          &mockWg,
	}
)

// Мок для parser.Parse
func mockParse(expression string, id int) {
	defer orchestratorState.Wg.Done()
	// Простая заглушка: записываем результат в Expressions
	orchestratorState.Expressions.Set(strconv.Itoa(id), entities.ClientResponse{Id: id, Status: "Done", Result: 42.0})
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
	getHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("GET")) })
	postHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("POST")) })

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
		// Сбрасываем состояние
		orchestratorState.Tasks = &mockTasks{queue: []entities.Task{{Id: 1, Arg1: 2, Arg2: 3, Operation: "+"}}}

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), logger)

		req := httptest.NewRequest("GET", "/internal/task", nil)
		w := httptest.NewRecorder()
		getHandler(ctx).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("getHandler status = %v, want %v", w.Code, http.StatusOK)
		}
		var task entities.Task
		json.Unmarshal(w.Body.Bytes(), &task)
		if task.Id != 1 {
			t.Errorf("getHandler task.Id = %v, want 1", task.Id)
		}
	})

	t.Run("No tasks", func(t *testing.T) {
		// Сбрасываем состояние
		orchestratorState.Tasks = &mockTasks{queue: []entities.Task{}}

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

func TestPostHandler(t *testing.T) {
	t.Run("Valid task result", func(t *testing.T) {
		// Сбрасываем состояние
		orchestratorState.ParsersTree = &mockParsersTree{data: make(map[int]*chan float64)}
		ch := make(chan float64, 1)
		orchestratorState.ParsersTree.Insert(1, &ch)

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), logger)

		taskRes := entities.ClientResponse{Id: 1, Result: 42.0}
		body, _ := json.Marshal(taskRes)
		req := httptest.NewRequest("POST", "/internal/task", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		postHandler(ctx).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("postHandler status = %v, want %v", w.Code, http.StatusOK)
		}
		select {
		case result := <-ch:
			if result != 42.0 {
				t.Errorf("postHandler result = %v, want 42.0", result)
			}
		default:
			t.Errorf("No result sent to channel")
		}
	})

	t.Run("Task not found", func(t *testing.T) {
		// Сбрасываем состояние
		orchestratorState.ParsersTree = &mockParsersTree{data: make(map[int]*chan float64)}

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), logger)

		taskRes := entities.ClientResponse{Id: 999, Result: 42.0}
		body, _ := json.Marshal(taskRes)
		req := httptest.NewRequest("POST", "/internal/task", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		postHandler(ctx).ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("postHandler status = %v, want %v", w.Code, http.StatusNotFound)
		}
	})
}

//func TestCalculateHandler(t *testing.T) {
//	// Заменяем parser.Parse на мок
//	defer func(orig func(string, int)) { parser.Parse = orig }(parser.Parse)
//	parser.Parse = mockParse
//
//	tests := []struct {
//		name       string
//		body       string
//		wantStatus int
//	}{
//		{name: "Valid expression", body: `{"expression": "2 + 3"}`, wantStatus: http.StatusCreated},
//		{name: "Invalid expression", body: `{"expression": "2 + a"}`, wantStatus: http.StatusUnprocessableEntity},
//		{name: "Invalid JSON", body: `{"expression": "2 + 3"`, wantStatus: http.StatusInternalServerError},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// Сбрасываем состояние
//			orchestratorState.Expressions = pkg.NewSafeMap()
//			orchestratorState.TasksLastID = &mockTasksLastID{value: 1}
//
//			var logBuf bytes.Buffer
//			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
//			ctx := logger2.WithLogger(context.Background(), logger)
//
//			req := httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer([]byte(tt.body)))
//			w := httptest.NewRecorder()
//			calculateHandler(ctx).ServeHTTP(w, req)
//
//			if w.Code != tt.wantStatus {
//				t.Errorf("calculateHandler status = %v, want %v", w.Code, tt.wantStatus)
//			}
//			if tt.wantStatus == http.StatusCreated {
//				var resp entities.ClientResponse
//				json.Unmarshal(w.Body.Bytes(), &resp)
//				if resp.Id != 1 {
//					t.Errorf("calculateHandler response Id = %v, want 1", resp.Id)
//				}
//			}
//		})
//	}
//}

func TestExpressionHandler(t *testing.T) {
	t.Run("Get all expressions", func(t *testing.T) {
		// Сбрасываем состояние
		orchestratorState.Expressions = pkg.NewSafeMap()
		orchestratorState.Expressions.Set("1", entities.ClientResponse{Id: 1, Status: "Done", Result: 10.0})
		orchestratorState.Expressions.Set("2", entities.ClientResponse{Id: 2, Status: "In progress"})

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), logger)

		req := httptest.NewRequest("GET", "/api/v1/expressions", nil)
		w := httptest.NewRecorder()
		expressionHandler(ctx).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expressionHandler status = %v, want %v", w.Code, http.StatusOK)
		}
		var resp struct {
			Expressions []entities.ClientResponse `json:"expressions"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Expressions) != 2 {
			t.Errorf("expressionHandler len(expressions) = %v, want 2", len(resp.Expressions))
		}
	})
}

func TestExpressionIDHandler(t *testing.T) {
	t.Run("Valid ID", func(t *testing.T) {
		// Сбрасываем состояние
		orchestratorState.Expressions = pkg.NewSafeMap()
		orchestratorState.Expressions.Set("1", entities.ClientResponse{Id: 1, Status: "Done", Result: 10.0})

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), logger)

		req := httptest.NewRequest("GET", "/api/v1/expressions/1", nil)
		w := httptest.NewRecorder()
		expressionIDHandler(ctx).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expressionIDHandler status = %v, want %v", w.Code, http.StatusOK)
		}
		var resp entities.ClientResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Id != 1 {
			t.Errorf("expressionIDHandler response Id = %v, want 1", resp.Id)
		}
	})

	t.Run("Invalid ID", func(t *testing.T) {
		// Сбрасываем состояние
		orchestratorState.Expressions = pkg.NewSafeMap()

		var logBuf bytes.Buffer
		logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
		ctx := logger2.WithLogger(context.Background(), logger)

		req := httptest.NewRequest("GET", "/api/v1/expressions/invalid", nil)
		w := httptest.NewRecorder()
		expressionIDHandler(ctx).ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expressionIDHandler status = %v, want %v", w.Code, http.StatusInternalServerError)
		}
	})
}
