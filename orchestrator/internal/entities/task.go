package entities

// Task is a struct that contains the task to be executed
type Task struct {
	Id            int     `json:"id,omitempty"`
	Arg1          float64 `json:"arg1,omitempty"`
	Arg2          float64 `json:"arg2,omitempty"`
	Operation     string  `json:"operation,omitempty"`
	OperationTime int     `json:"operation_time,omitempty"`
}
