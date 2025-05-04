package entities

// ClientRequest is a struct that contains the request from the client
type ClientRequest struct {
	Expression string `json:"expression"`
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
