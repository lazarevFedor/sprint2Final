package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	obj "orchestrator/internal/entities"
	"orchestrator/internal/parser"
	"regexp"
)

//TODO: add comments, api/v1/expressions/:id endpoint, reset done tasks with goroutine with ticker

// isValidExpression checks if the expression is valid
func isValidExpression(expression string) bool {
	re := regexp.MustCompile("^[\\d+\\-*/\\s]+$")
	return re.MatchString(expression)
}

func agentMiddleware(getHandler, postHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getHandler.ServeHTTP(w, r)
		case http.MethodPost:
			postHandler.ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if obj.Tasks.IsEmpty() {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	task := obj.Tasks.Dequeue()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	var taskRes obj.ClientResponse
	if err := json.NewDecoder(r.Body).Decode(&taskRes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	node := obj.ParsersTree.Search(taskRes.Id)
	// TODO: invalid data, code 422?
	if node == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	ch := node.Value.(*chan float64)
	*ch <- taskRes.Result
	w.WriteHeader(http.StatusOK)
}

// calculateHandler handles the /api/v1/calculate endpoint
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var clientRequest obj.ClientRequest
	var clientResponse obj.ClientResponse
	err := json.NewDecoder(r.Body).Decode(&clientRequest)
	if err != nil {
		clientResponse.Error = errors.New("Internal server error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !isValidExpression(clientRequest.Expression) {
		clientResponse.Error = errors.New("Expression is not valid")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	clientResponse.Id = obj.TasksLastID.GetValue()
	obj.TasksLastID.Increment()
	obj.Wg.Add(1)
	go parser.Parse(clientRequest.Expression, clientResponse.Id)
}

func expressionHandler(w http.ResponseWriter, r *http.Request) {
	expressionsMap := obj.Expressions.GetAll()
	expressions := make([]obj.ClientResponse, 0, len(expressionsMap))

	for _, expr := range expressionsMap {
		task, ok := expr.(obj.ClientResponse)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		expressions = append(expressions, task)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{"expressions": expressions}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// StartServer starts the server on port 8080 and listens for incoming requests
func StartServer() error {
	//obj.TasksLastID.Increment()
	mux := http.NewServeMux()
	// Handle functions for client requests
	mux.HandleFunc("/api/v1/calculate", calculateHandler)
	mux.HandleFunc("/api/v1/expressions", expressionHandler)
	//mux.HandleFunc("/api/v1/expressions/:id", nil)
	// Handle functions for agent requests
	mux.Handle("/internal/task", agentMiddleware(http.HandlerFunc(getHandler), http.HandlerFunc(postHandler)))
	// Start the server
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return fmt.Errorf("could not start server: %v", err)
	}
	obj.Wg.Wait()
	return nil
}
