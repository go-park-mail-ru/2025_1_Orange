package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Auth interface {
<<<<<<< HEAD
	Logout(context.Context, string) error
	LogoutAll(context.Context, int, string) error
	GetUserIDBySession(context.Context, string) (int, string, error)
	CreateSession(context.Context, int, string) (string, error)
=======
	Logout(string) error
	LogoutAll(int, string) error
	GetUserIDBySession(string) (int, string, error)
	CreateSession(int, string) (string, error)
>>>>>>> c773955 (Made vacansies usecases and handlers)
	EmailExists(context.Context, string) (*dto.EmailExistsResponse, error)
}
