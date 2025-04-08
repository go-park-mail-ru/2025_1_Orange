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
}
