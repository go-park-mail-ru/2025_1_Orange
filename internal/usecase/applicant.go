package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Applicant interface {
	Register(context.Context, *dto.ApplicantRegister) (int, error)
	Login(context.Context, *dto.ApplicantLogin) (int, error)
	GetUser(context.Context, int) (*dto.ApplicantProfile, error)
}
