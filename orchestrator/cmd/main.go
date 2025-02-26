package main

import (
	"orchestrator/internal/server"
)

func main() {
	//pkg.Init()
	err := server.StartServer()
	if err != nil {
		return
	}
}
