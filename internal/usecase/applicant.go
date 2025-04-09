package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Applicant interface {
	Register(context.Context, *dto.ApplicantRegister) (int, error)
	Login(context.Context, *dto.ApplicantLogin) (int, error)
	GetUser(context.Context, int) (*dto.ApplicantProfileResponse, error)
	UpdateProfile(context.Context, int, *dto.ApplicantProfileUpdate) error
	UpdateAvatar(context.Context, int, int) error
}
