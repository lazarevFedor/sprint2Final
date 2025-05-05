package entities

// ClientResponse is a struct that contains the response to the client
type ClientResponse struct {
	userId int
	Id     int     `json:"id"`
	Status string  `json:"status,omitempty"`
	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}

type LoginResponse struct {
	Token string
}

func (cr *ClientResponse) GetUserId() int {
	return cr.userId
}

func (cr *ClientResponse) SetUserId(userId int) {
	cr.userId = userId
}
