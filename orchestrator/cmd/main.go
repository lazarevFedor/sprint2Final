package main

import (
	"orchestrator/internal/server"
	"pkg/logger"
)

func main() {
	err := server.StartServer()
	logger.Init()
	if err != nil {
		return
	}
}
