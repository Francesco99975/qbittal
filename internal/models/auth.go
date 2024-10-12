package models

type LoginPayload struct {
	Server   string `json:"server"`
	Password string `json:"password"`
}

type LoginInfo struct {
	Token string `json:"token"`
}

type TokenInfo struct {
	Token string `json:"token"`
}

type CheckResponse struct {
	Valid bool `json:"valid"`
}
