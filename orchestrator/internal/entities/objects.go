package entities

import (
	"pkg"
	"sync"
)

// TODO: Add a comment to this block
var (
	Wg          = &sync.WaitGroup{}
	ParserMutex = &sync.Mutex{}
	ParsersTree = pkg.NewRBTree()
	Tasks       = &pkg.Queue{}
	Expressions = pkg.NewSafeMap()
	TasksLastID = &pkg.Counter{}
)

// ClientRequest is a struct that contains the request from the client
type ClientRequest struct {
	Expression string `json:"expression,omitempty"`
}

// ClientResponse is a struct that contains the response to the client
type ClientResponse struct {
	Id     int     `json:"id"`
	Status string  `json:"status,omitempty"`
	Result float64 `json:"result,omitempty"`
	Error  error   `json:"error,omitempty"`
}

// Task is a struct that contains the task to be executed
type Task struct {
	Id            int     `json:"id,omitempty"`
	Arg1          float64 `json:"arg1,omitempty"`
	Arg2          float64 `json:"arg2,omitempty"`
	Operation     string  `json:"operation,omitempty"`
	OperationTime int     `json:"operation_time,omitempty"`
}
