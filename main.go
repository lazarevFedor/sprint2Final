package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sprint1Final/Calculator"
)

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	var resp Response
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp.Error = "Internal server error"
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	result, err := Calculator.Calc(req.Expression)
	if err != nil {
		if err.Error() == "wrong symbol" {
			resp.Error = "Expression is not valid"
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
		resp.Error = "Internal server error"
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		resp.Result = fmt.Sprintf("%f", result)
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		resp.Error = "Internal server error"
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/api/v1/calculate", calculateHandler)
	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
