package entities

// ClientResponse is a struct that contains the response to the client
type ClientResponse struct {
	Id     int     `json:"id"`
	Status string  `json:"status,omitempty"`
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}
