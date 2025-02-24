package main

import (
	"orchestrator/internal/server"
	"pkg"
)

func main() {
	err := server.StartServer()
	pkg.Init()
	if err != nil {
		return
	}
}
