package repository

import (
	"ResuMatch/internal/entity"
	"context"
)

type PollRepository interface {
	CreateVote(ctx context.Context, vote *entity.Vote) error
	GetAll(ctx context.Context) ([]*entity.Poll, error)
	GetVotesByPoll(ctx context.Context, pollID int) ([]*entity.VoteStats, error)
	GetNewPoll(ctx context.Context, userID int, role string) (*entity.Poll, error)
}
