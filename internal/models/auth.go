package models

type LoginPayload struct {
	Server   string `json:"server"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type LoginInfo struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

type TokenInfo struct {
	Token string `json:"token"`
}

type CheckResponse struct {
	Valid bool `json:"valid"`
}
