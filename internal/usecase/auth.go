package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Auth interface {
	Logout(string) error
	LogoutAll(int, string) error
	GetUserIDBySession(string) (int, string, error)
	CreateSession(int, string) (string, error)
	EmailExists(context.Context, string) (*dto.EmailExistsResponse, error)
}
