package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type PollRepository interface {
	CreateVote(ctx context.Context, vote *entity.Vote) error
}
