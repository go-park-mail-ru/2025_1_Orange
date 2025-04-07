package dto

type ApplicantRegister struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,max=30"`
	LastName  string `json:"last_name" validate:"required,max=30"`
}

type ApplicantLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type EmployerRegister struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	CompanyName  string `json:"company_name" validate:"required,max=64"`
	LegalAddress string `json:"legal_address,omitempty" validate:"max=255"`
}

type EmployerLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
