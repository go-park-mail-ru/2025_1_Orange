package auth

import (
	"ResuMatch/internal/metrics"
	authPROTO "ResuMatch/internal/transport/grpc/auth/proto"
	"ResuMatch/internal/transport/grpc/interceptors"
	"context"
	"github.com/prometheus/client_golang/prometheus"
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
		grpc.WithUnaryInterceptor(interceptors.RequestIDClientInterceptor()),
	)
	if err != nil {
		return nil, err
	}

	authClient := authPROTO.NewAuthServiceClient(grpcConn)

	return &Gateway{authClient: authClient}, nil
}

func (gw *Gateway) Logout(ctx context.Context, session string) error {
	timer := prometheus.NewTimer(metrics.AuthServiceCallDuration.WithLabelValues("Logout"))
	defer timer.ObserveDuration()

	_, err := gw.authClient.Logout(ctx, &authPROTO.LogoutRequest{Session: session})
	if err != nil {
		metrics.AuthServiceCallCounter.WithLabelValues("Logout", "500").Inc()
		return err
	}

	metrics.AuthServiceCallCounter.WithLabelValues("Logout", "200").Inc()
	return nil
}

func (gw *Gateway) LogoutAll(ctx context.Context, userID int, role string) error {
	timer := prometheus.NewTimer(metrics.AuthServiceCallDuration.WithLabelValues("LogoutAll"))
	defer timer.ObserveDuration()

	_, err := gw.authClient.LogoutAll(ctx, &authPROTO.LogoutAllRequest{UserId: uint64(userID), Role: role})
	if err != nil {
		metrics.AuthServiceCallCounter.WithLabelValues("LogoutAll", "500").Inc()
		return err
	}

	metrics.AuthServiceCallCounter.WithLabelValues("LogoutAll", "200").Inc()
	return nil
}

func (gw *Gateway) GetUserIDBySession(ctx context.Context, session string) (int, string, error) {
	timer := prometheus.NewTimer(metrics.AuthServiceCallDuration.WithLabelValues("GetUserIDBySession"))
	defer timer.ObserveDuration()

	resp, err := gw.authClient.GetUserIDBySession(ctx, &authPROTO.GetUserIDBySessionRequest{Session: session})
	if err != nil {
		metrics.AuthServiceCallCounter.WithLabelValues("GetUserIDBySession", "500").Inc()
		return -1, "", err
	}

	metrics.AuthServiceCallCounter.WithLabelValues("GetUserIDBySession", "200").Inc()
	return int(resp.UserId), resp.Role, nil
}

func (gw *Gateway) CreateSession(ctx context.Context, userID int, role string) (string, error) {
	timer := prometheus.NewTimer(metrics.AuthServiceCallDuration.WithLabelValues("CreateSession"))
	defer timer.ObserveDuration()

	resp, err := gw.authClient.CreateSession(ctx, &authPROTO.CreateSessionRequest{UserId: uint64(userID), Role: role})
	if err != nil {
		metrics.AuthServiceCallCounter.WithLabelValues("CreateSession", "500").Inc()
		return "", err
	}
	metrics.AuthServiceCallCounter.WithLabelValues("CreateSession", "200").Inc()
	return resp.Session, nil
}
