package usecase

import (
	"context"
)

type Auth interface {
	Logout(ctx context.Context, session string) error
	LogoutAll(ctx context.Context, userID int, role string) error
	GetUserIDBySession(ctx context.Context, session string) (int, string, error)
	CreateSession(ctx context.Context, userID int, role string) (string, error)
}
