package main

import "orchestrator/internal/server"

func main() {
	err := server.StartServer()
	if err != nil {
		return
	}
}
