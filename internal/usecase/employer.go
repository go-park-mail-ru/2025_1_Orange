package usecase

import "ResuMatch/internal/entity/dto"

type Employer interface {
	Register(*dto.EmployerRegister) (int, error)
	Login(*dto.EmployerLogin) (int, error)
	GetUser(employerID int) (*dto.EmployerProfile, error)
}
