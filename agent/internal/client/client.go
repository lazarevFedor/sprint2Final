package client

import (
	"agent/internal/demon"
	"agent/internal/entities"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	logger2 "pkg/logger"
	"strconv"
	"time"
)

// ManageTasks is a function that manages tasks
func ManageTasks(ctx context.Context) {
	logger := logger2.GetLogger(ctx)
	logger.Info("ManageTasks: Start")
	client := &http.Client{}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	computingPower, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if err != nil || computingPower <= 0 {
		computingPower = 1
	}

	taskChan := make(chan entities.AgentResponse, 1)

	for i := 0; i < computingPower; i++ {
		go worker(client, taskChan, ctx)
	}

	for range ticker.C {
		req, err := http.NewRequest("GET", "http://orchestrator:8080/internal/task", nil)
		if err != nil {
			logger.Error("ManageTasks: GET request error:", "err", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode == http.StatusNotFound {
			continue
		} else if resp.StatusCode == http.StatusInternalServerError {
			logger.Error("ManageTasks: getting response error:", "err", err)
			return
		}

		var task entities.AgentResponse
		if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
			logger.Error("ManageTasks: decoding response error:", "err", err)
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			logger.Error("ManageTasks: closing response error:", "err", err)
			return
		}
		taskChan <- task
		logger.Info("ManageTasks: Task received")
	}
}

// worker is a function that processes tasks
func worker(client *http.Client, taskChan <-chan entities.AgentResponse, ctx context.Context) {
	for task := range taskChan {
		solveTask(client, task, ctx)
	}
}

// solveTask is a function that solves the task
func solveTask(client *http.Client, task entities.AgentResponse, ctx context.Context) {
	logger := logger2.GetLogger(ctx)
	result, err := demon.CalculateExpression(task.Arg1, task.Arg2, task.Operation, task.OperationTime)
	if err != nil {
		logger.Error("solveTask: calculating expression error:", "err", err)
		return
	}
	request := entities.AgentRequest{}
	request.Id = task.Id
	request.Result = result
	data, err := json.Marshal(request)
	if err != nil {
		logger.Error("solveTask: marshalling request error:", "err", err)
		return
	}

	req, err := http.NewRequest("POST", "http://orchestrator:8080/internal/task", bytes.NewBuffer(data))
	if err != nil {
		logger.Error("solveTask: POST request error:", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Error("solveTask: sending response error:", "err", err)
		return
	}
	err = resp.Body.Close()
	if err != nil {
		logger.Error("solveTask: closing response error:", "err", err)
		return
	}
	logger.Info("solveTask: Task solved")
}
