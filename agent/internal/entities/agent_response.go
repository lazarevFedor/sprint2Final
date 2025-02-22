package entities

type AgentResponse struct {
	Id            int     `json:"id,omitempty"`
	Arg1          float64 `json:"arg1,omitempty"`
	Arg2          float64 `json:"arg2,omitempty"`
	Operation     string  `json:"operation,omitempty"`
	OperationTime int     `json:"operation_time,omitempty"`
}
