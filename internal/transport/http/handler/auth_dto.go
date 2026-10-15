package handler

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}
