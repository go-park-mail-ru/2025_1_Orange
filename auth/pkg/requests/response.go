package requests




type AuthCheckResponse struct {
	Login string `json:"login"`
	Role  string `json:"role"`
}