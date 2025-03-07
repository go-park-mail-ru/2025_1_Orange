package requests

type (
	SignupRequest struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
		RepeatPassword string `json:"repeatPassword"`
		FirstName      string `json:"firstName"`
		LastName       string `json:"lastName"`
		CompanyName    string `json:"companyName"`
		CompanyAddress string `json:"companyAddress"`
	}

	SigninRequest struct {
		Email    string `json:"login"`
		Password string `json:"password"`
	}
)
