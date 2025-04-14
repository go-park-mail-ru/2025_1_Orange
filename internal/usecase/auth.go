package usecase

import (
	"ResuMatch/internal/entity/dto"
	"context"
)

type Auth interface {
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> a6396a4 (Fix mistakes)
	Logout(context.Context, string) error
	LogoutAll(context.Context, int, string) error
	GetUserIDBySession(context.Context, string) (int, string, error)
	CreateSession(context.Context, int, string) (string, error)
<<<<<<< HEAD
=======
	Logout(string) error
	LogoutAll(int, string) error
	GetUserIDBySession(string) (int, string, error)
	CreateSession(int, string) (string, error)
>>>>>>> c773955 (Made vacansies usecases and handlers)
=======
>>>>>>> a6396a4 (Fix mistakes)
	EmailExists(context.Context, string) (*dto.EmailExistsResponse, error)
}
