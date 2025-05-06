package auth

import (
	authPROTO "ResuMatch/internal/transport/grpc/auth/proto"
	"ResuMatch/internal/transport/grpc/utils"
	"ResuMatch/internal/usecase"
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPC struct {
	authPROTO.UnimplementedAuthServiceServer
	authUC usecase.Auth
}

func NewGRPC(authUC usecase.Auth) *GRPC {
	return &GRPC{
		authUC: authUC,
	}
}

func (service *GRPC) Logout(ctx context.Context, request *authPROTO.LogoutRequest) (*emptypb.Empty, error) {
	err := service.authUC.Logout(ctx, request.Session)
	if err != nil {
		return nil, utils.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (service *GRPC) LogoutAll(ctx context.Context, request *authPROTO.LogoutAllRequest) (*emptypb.Empty, error) {
	err := service.authUC.LogoutAll(ctx, int(request.UserId), request.Role)
	if err != nil {
		return nil, utils.ToGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (service *GRPC) GetUserIDBySession(ctx context.Context, request *authPROTO.GetUserIDBySessionRequest) (*authPROTO.GetUserIDBySessionResponse, error) {
	userID, role, err := service.authUC.GetUserIDBySession(ctx, request.Session)
	if err != nil {
		return nil, utils.ToGRPCError(err)
	}
	return &authPROTO.GetUserIDBySessionResponse{
		UserId: uint64(userID),
		Role:   role,
	}, nil
}

func (service *GRPC) CreateSession(ctx context.Context, request *authPROTO.CreateSessionRequest) (*authPROTO.CreateSessionResponse, error) {
	session, err := service.authUC.CreateSession(ctx, int(request.UserId), request.Role)
	if err != nil {
		return nil, utils.ToGRPCError(err)
	}
	return &authPROTO.CreateSessionResponse{
		Session: session,
	}, nil
}
