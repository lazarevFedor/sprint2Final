package entities

import "time"

// ClientResponse is a struct that contains the response to the client
type ClientResponse struct {
	Id        int     `json:"id"`
	Status    string  `json:"status,omitempty"`
	Result    float64 `json:"result,omitempty"`
	Error     string  `json:"error,omitempty"`
	timestamp time.Time
}

// SetTimestamp sets the timestamp for the ClientResponse
func (cr *ClientResponse) SetTimestamp(t time.Time) {
	cr.timestamp = t
}

// GetTimestamp gets the timestamp for the ClientResponse
func (cr *ClientResponse) GetTimestamp() time.Time {
	return cr.timestamp
}
