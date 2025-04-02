package dto

type EmployerProfile struct {
	ID           int    `json:"id"`
	CompanyName  string `json:"company_name"`
	Email        string `json:"email"`
	LegalAddress string `json:"legal_address,omitempty"`
}
