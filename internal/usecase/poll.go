package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Poll interface {
	Vote(ctx context.Context, userID int, role string, voteDTO *dto.VotePollRequest) error
}
