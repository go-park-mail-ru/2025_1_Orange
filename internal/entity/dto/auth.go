package dto

type AuthCredentials struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ApplicantRegister struct {
	AuthCredentials
	FirstName string `json:"first_name" validate:"required,max=30"`
	LastName  string `json:"last_name" validate:"required,max=30"`
}

type Login struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type EmployerRegister struct {
	AuthCredentials
	CompanyName  string `json:"company_name" validate:"required,max=64"`
	LegalAddress string `json:"legal_address,omitempty" validate:"max=255"`
}

type AuthResponse struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

type EmailExistsRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type EmailExistsResponse struct {
	Exists bool   `json:"exists"`
	Role   string `json:"role"`
}

type AuthResponse struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}
