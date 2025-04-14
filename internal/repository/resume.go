package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type ResumeRepository interface {
	Create(ctx context.Context, resume *entity.Resume) (*entity.Resume, error)
	AddSkills(ctx context.Context, resumeID int, skillIDs []int) error
	AddSpecializations(ctx context.Context, resumeID int, specializationIDs []int) error
	AddWorkExperience(ctx context.Context, workExperience *entity.WorkExperience) (*entity.WorkExperience, error)
	GetByID(ctx context.Context, id int) (*entity.Resume, error)
	GetSkillsByResumeID(ctx context.Context, resumeID int) ([]entity.Skill, error)
	GetSpecializationsByResumeID(ctx context.Context, resumeID int) ([]entity.Specialization, error)
	GetWorkExperienceByResumeID(ctx context.Context, resumeID int) ([]entity.WorkExperience, error)
	Update(ctx context.Context, resume *entity.Resume) (*entity.Resume, error)
	Delete(ctx context.Context, id int) error
	DeleteSkills(ctx context.Context, resumeID int) error
	DeleteSpecializations(ctx context.Context, resumeID int) error
	DeleteWorkExperiences(ctx context.Context, resumeID int) error
	UpdateWorkExperience(ctx context.Context, workExperience *entity.WorkExperience) (*entity.WorkExperience, error)
	DeleteWorkExperience(ctx context.Context, id int) error
	GetAll(ctx context.Context) ([]entity.Resume, error)
	GetAllResumesByApplicantID(ctx context.Context, applicantID int) ([]entity.Resume, error)
	FindSkillIDsByNames(ctx context.Context, skillNames []string) ([]int, error)
	FindSpecializationIDByName(ctx context.Context, specializationName string) (int, error)
	FindSpecializationIDsByNames(ctx context.Context, specializationNames []string) ([]int, error)
	CreateSkillIfNotExists(ctx context.Context, skillName string) (int, error)
	CreateSpecializationIfNotExists(ctx context.Context, specializationName string) (int, error)
}
