package repository

import "context"

type SessionRepository interface {
	CreateSession(ctx context.Context, userID int, role string) (string, error)
	GetSession(ctx context.Context, sessionToken string) (userID int, role string, err error)
	DeleteSession(ctx context.Context, sessionToken string) error
	DeleteAllSessions(ctx context.Context, userID int, role string) error
}
