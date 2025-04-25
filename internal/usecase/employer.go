package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Employer interface {
	Register(context.Context, *dto.EmployerRegister) (int, error)
	Login(context.Context, *dto.Login) (int, error)
	GetUser(context.Context, int) (*dto.EmployerProfileResponse, error)
	UpdateProfile(context.Context, int, *dto.EmployerProfileUpdate) error
	UpdateLogo(context.Context, int, int) error
	EmailExists(context.Context, string) (*dto.EmailExistsResponse, error)
}
