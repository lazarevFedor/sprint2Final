package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sprint1Final/internal/calc"
)

// Request struct is used to parse the request body
type Request struct {
	Expression string `json:"expression"`
}

// Response struct is used to send the response
type Response struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// calculateHandler is a handler function that is called when a request is made to /api/v1/calculate
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	var resp Response

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error = "Internal server error"
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	result, err := calc.Calc(req.Expression)
	if err != nil {
		if err.Error() == "wrong symbol" {
			resp.Error = "Expression is not valid"
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			resp.Error = "Internal server error"
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		resp.Result = fmt.Sprintf("%.2f", result)
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		resp.Error = "Internal server error"
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// StartServer starts the server on port 8080
func StartServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/calculate", calculateHandler)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		return fmt.Errorf("could not start server: %v", err)
	}
	return nil
}
