package requests

//easyjson:json
type (
	SignupRequest struct {
		Login     string `json:"login"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		BirthDate string `json:"birth_date"`
		Name      string `json:"name"`
	}

	SigninRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
)
