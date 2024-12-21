package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	calc "sprint1Final/Calculator"
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
		http.Error(w, fmt.Sprintf("Что-то пошло не так"), http.StatusInternalServerError)
		return
	}
	result, err := calc.Calc(req.Expression)
	if err != nil {
		resp.Error = "Expression is not valid"
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		resp.Result = fmt.Sprintf("%f", result)
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, fmt.Sprintf("Что-то пошло не так"), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/api/v1/calculate", calculateHandler)
	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
