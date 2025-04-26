package poll

import (
	"ResuMatch/internal/entity/dto"
	pollPROTO "ResuMatch/internal/transport/grpc/poll/proto"
	"ResuMatch/internal/usecase"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPC struct {
	pollPROTO.UnimplementedPollServiceServer
	pollUC usecase.Poll
}

func NewGRPC(pollUC usecase.Poll) *GRPC {
	return &GRPC{
		pollUC: pollUC,
	}
}
func (s *GRPC) Vote(ctx context.Context, req *pollPROTO.VoteRequest) (*emptypb.Empty, error) {
	voteDTO := &dto.VotePollRequest{
		PollID: int(req.VoteData.PollId),
		Answer: int(req.VoteData.Answer),
	}
	err := s.pollUC.Vote(ctx, int(req.UserId), req.Role, voteDTO)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *GRPC) GetStats(ctx context.Context, _ *emptypb.Empty) (*pollPROTO.GetStatsResponse, error) {
	stats, err := s.pollUC.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	resp := &pollPROTO.GetStatsResponse{
		Polls: make([]*pollPROTO.PollStatsResponse, 0, len(stats)),
	}

	for _, stat := range stats {
		pollStat := &pollPROTO.PollStatsResponse{
			PollId:  uint64(stat.PollID),
			Average: stat.Average,
			Name:    stat.Name,
			Stars:   make([]*pollPROTO.StarStats, 0, len(stat.Stars)),
		}

		for _, star := range stat.Stars {
			pollStat.Stars = append(pollStat.Stars, &pollPROTO.StarStats{
				Star:       uint64(star.Star),
				Amount:     uint64(star.Amount),
				Percentage: star.Percentage,
			})
		}

		resp.Polls = append(resp.Polls, pollStat)
	}

	return resp, nil
}

func (s *GRPC) GetNewPoll(ctx context.Context, req *pollPROTO.GetNewPollRequest) (*pollPROTO.PollResponse, error) {
	poll, err := s.pollUC.GetNewPoll(ctx, int(req.UserId), req.Role)
	if err != nil {
		return nil, err
	}

	return &pollPROTO.PollResponse{
		PollId: uint64(poll.PollID),
		Name:   poll.Name,
	}, nil
}
