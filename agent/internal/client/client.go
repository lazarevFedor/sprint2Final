package client

import (
	"agent/internal/demon"
	"agent/internal/entities"
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

//TODO: Add environment variable for computing power and operation time

// ManageTasks is a function that manages tasks
func ManageTasks() {
	client := &http.Client{}
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	//computingPower, err := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	//if err != nil || computingPower <= 0 {
	//	computingPower = 1
	//}

	taskChan := make(chan entities.AgentResponse, 1)

	for i := 0; i < 1; i++ {
		go worker(client, taskChan)
	}

	for range ticker.C {
		req, err := http.NewRequest("GET", "http://localhost:8080/internal/task", nil)
		if err != nil {
			// Обработка ошибки
			continue
		}

		resp, err := client.Do(req)
		if err != nil || resp.StatusCode == http.StatusNotFound {
			// Обработка ошибки
			continue
		} else if resp.StatusCode == http.StatusInternalServerError {
			return
		}

		var task entities.AgentResponse
		if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
			// Обработка ошибки
			continue
		}
		err = resp.Body.Close()
		if err != nil {
			return
		}
		taskChan <- task
	}
}

// worker is a function that processes tasks
func worker(client *http.Client, taskChan <-chan entities.AgentResponse) {
	for task := range taskChan {
		solveTask(client, task)
	}
}

// solveTask is a function that solves the task
func solveTask(client *http.Client, task entities.AgentResponse) {
	result, err := demon.CalculateExpression(task.Arg1, task.Arg2, task.Operation, task.OperationTime)
	if err != nil {
		// Обработка ошибки
		return
	}
	request := entities.AgentRequest{}
	request.Id = task.Id
	request.Result = result
	data, err := json.Marshal(request)
	if err != nil {
		// Обработка ошибки
		return
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/internal/task", bytes.NewBuffer(data))
	if err != nil {
		// Обработка ошибки
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		// Обработка ошибки
		return
	}
	resp.Body.Close()
}
