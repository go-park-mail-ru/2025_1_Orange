package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type SkillRepository interface {
	GetByIDs(ctx context.Context, ids []int) ([]entity.Skill, error)
}
