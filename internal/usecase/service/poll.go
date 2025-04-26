package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
	"fmt"
	"github.com/asaskevich/govalidator"
)

type PollService struct {
	pollRepository repository.PollRepository
}

func NewPollService(pollRepository repository.PollRepository) usecase.Poll {
	return &PollService{pollRepository: pollRepository}
}

func (s *PollService) Vote(ctx context.Context, userID int, role string, voteDTO *dto.VotePollRequest) error {
	if isValid, err := govalidator.ValidateStruct(voteDTO); !isValid {
		return entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("неправильный формат данных: %w", err),
		)
	}

	voteEntity := &entity.Vote{
		PollID: voteDTO.PollID,
		UserID: userID,
		Role:   role,
		Answer: voteDTO.Answer,
	}

	err := s.pollRepository.CreateVote(ctx, voteEntity)
	if err != nil {
		return err
	}

	return nil
}
