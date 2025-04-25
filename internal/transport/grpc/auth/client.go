package auth

import (
	authPROTO "ResuMatch/internal/transport/grpc/auth/proto"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	authClient authPROTO.AuthServiceClient
}

func NewGateway(connectAddr string) (*Gateway, error) {
	grpcConn, err := grpc.NewClient(
		connectAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	authClient := authPROTO.NewAuthServiceClient(grpcConn)

	return &Gateway{authClient: authClient}, nil
}

func (gw *Gateway) Logout(ctx context.Context, session string) error {
	_, err := gw.authClient.Logout(ctx, &authPROTO.LogoutRequest{Session: session})
	if err != nil {
		return err
	}
	return nil
}

func (gw *Gateway) LogoutAll(ctx context.Context, userID int, role string) error {
	_, err := gw.authClient.LogoutAll(ctx, &authPROTO.LogoutAllRequest{UserId: uint64(userID), Role: role})
	if err != nil {
		return err
	}
	return nil
}

func (gw *Gateway) GetUserIDBySession(ctx context.Context, session string) (int, string, error) {
	resp, err := gw.authClient.GetUserIDBySession(ctx, &authPROTO.GetUserIDBySessionRequest{Session: session})
	if err != nil {
		return -1, "", err
	}
	return int(resp.UserId), resp.Role, nil
}

func (gw *Gateway) CreateSession(ctx context.Context, userID int, role string) (string, error) {
	resp, err := gw.authClient.CreateSession(ctx, &authPROTO.CreateSessionRequest{UserId: uint64(userID), Role: role})
	if err != nil {
		return "", err
	}
	return resp.Session, nil
}
