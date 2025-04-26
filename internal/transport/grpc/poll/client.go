package poll

import (
	"ResuMatch/internal/entity/dto"
	pollPROTO "ResuMatch/internal/transport/grpc/poll/proto"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Gateway struct {
	pollClient pollPROTO.PollServiceClient
}

func NewGateway(connectAddr string) (*Gateway, error) {
	grpcConn, err := grpc.NewClient(
		connectAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	pollClient := pollPROTO.NewPollServiceClient(grpcConn)

	return &Gateway{pollClient: pollClient}, nil
}

func (gw *Gateway) Vote(ctx context.Context, userID int, role string, voteDTO *dto.VotePollRequest) error {
	_, err := gw.pollClient.Vote(ctx, &pollPROTO.VoteRequest{
		UserId: uint64(userID),
		Role:   role,
		VoteData: &pollPROTO.VotePollRequest{
			PollId: uint64(voteDTO.PollID),
			Answer: uint64(voteDTO.Answer),
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (gw *Gateway) GetStats(ctx context.Context) ([]*dto.PollStatsResponse, error) {
	resp, err := gw.pollClient.GetStats(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	stats := make([]*dto.PollStatsResponse, 0, len(resp.Polls))
	for _, poll := range resp.Polls {
		stars := make([]*dto.StarStats, 0, len(poll.Stars))
		for _, star := range poll.Stars {
			stars = append(stars, &dto.StarStats{
				Star:       int(star.Star),
				Amount:     int(star.Amount),
				Percentage: star.Percentage,
			})
		}

		stats = append(stats, &dto.PollStatsResponse{
			PollID:  int(poll.PollId),
			Average: poll.Average,
			Name:    poll.Name,
			Stars:   stars,
		})
	}

	return stats, nil
}

func (gw *Gateway) GetNewPoll(ctx context.Context, userID int, role string) (*dto.PollResponse, error) {
	resp, err := gw.pollClient.GetNewPoll(ctx, &pollPROTO.GetNewPollRequest{
		UserId: uint64(userID),
		Role:   role,
	})

	if err != nil {
		return nil, err
	}

	return &dto.PollResponse{
		PollID: int(resp.PollId),
		Name:   resp.Name,
	}, nil
}
