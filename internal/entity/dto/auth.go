package dto

type AuthCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ApplicantRegister struct {
	AuthCredentials
	FirstName string `json:"first_name" valid:"runelength(2|30)"`
	LastName  string `json:"last_name" valid:"runelength(2|30)"`
}

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type EmployerRegister struct {
	AuthCredentials
	CompanyName  string `json:"company_name" valid:"runelength(2|30)"`
	LegalAddress string `json:"legal_address,omitempty" valid:"runelength(5|100)"`
}

type AuthResponse struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

type EmailExistsRequest struct {
	Email string `json:"email"`
}

type EmailExistsResponse struct {
	Exists bool   `json:"exists"`
	Role   string `json:"role"`
}
