package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Auth interface {
	Logout(context.Context, string) error
	LogoutAll(context.Context, int, string) error
	GetUserIDBySession(context.Context, string) (int, string, error)
	CreateSession(context.Context, int, string) (string, error)
	EmailExists(context.Context, string) (*dto.EmailExistsResponse, error)
}
