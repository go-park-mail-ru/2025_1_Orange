package usecase

import "ResuMatch/internal/entity/dto"

type Applicant interface {
	Register(*dto.ApplicantRegister) (int, error)
	Login(*dto.ApplicantLogin) (int, error)
	GetUser(applicantID int) (*dto.ApplicantProfile, error)
}
