package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
	"mime/multipart"
)

type Applicant interface {
	Register(context.Context, *dto.ApplicantRegister) (int, error)
	Login(context.Context, *dto.ApplicantLogin) (int, error)
	GetUser(context.Context, int) (*dto.ApplicantProfileResponse, error)
	UpdateProfile(context.Context, int, *dto.ApplicantProfileUpdate) error
	UploadAvatar(context.Context, int, *multipart.FileHeader) (*dto.UploadAvatarResponse, error)
}
