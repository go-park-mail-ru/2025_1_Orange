package delivery_auth_grpc

import (
	"auth/repository/profile"
	"auth/repository/session"

	"google.golang.org/grpc"
)

type authGrpc struct {
	grpcServ *grpc.Server
}

type server struct {
	userRepo    *profile.RepoPostgre
	sessionRepo *session.SessionRepo
}
