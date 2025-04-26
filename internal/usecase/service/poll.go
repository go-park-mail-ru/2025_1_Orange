package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
	"fmt"
	"github.com/asaskevich/govalidator"
	"math"
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

func (s *PollService) GetStats(ctx context.Context) ([]*dto.PollStatsResponse, error) {
	polls, err := s.pollRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var statsArray []*dto.PollStatsResponse

	for _, poll := range polls {
		stats, err := s.pollRepository.GetVotesByPoll(ctx, poll.ID)
		if err != nil {
			return nil, err
		}

		totalVotes := 0
		sum := 0
		starCounts := make(map[int]int)

		for _, stat := range stats {
			totalVotes += stat.Count
			sum += stat.Answer * stat.Count
			starCounts[stat.Answer] = stat.Count
		}

		var average float64
		if totalVotes > 0 {
			average = float64(sum) / float64(totalVotes)
			average = math.Round(average*10) / 10
		}

		var stars []*dto.StarStats
		for star := 1; star <= 5; star++ {
			count := starCounts[star]
			var percentage float64
			if totalVotes > 0 {
				percentage = math.Round(float64(count) / float64(totalVotes) * 100)
			}

			stars = append(stars, &dto.StarStats{
				Star:       star,
				Amount:     count,
				Percentage: percentage,
			})
		}

		statsArray = append(statsArray, &dto.PollStatsResponse{
			PollID:  poll.ID,
			Name:    poll.Name,
			Average: average,
			Stars:   stars,
		})
	}

	return statsArray, nil
}

func (s *PollService) GetNewPoll(ctx context.Context, userID int, role string) (*dto.PollResponse, error) {
	poll, err := s.pollRepository.GetNewPoll(ctx, userID, role)
	if err != nil {
		return nil, err
	}

	return &dto.PollResponse{
		PollID: poll.ID,
		Name:   poll.Name,
	}, nil
}
