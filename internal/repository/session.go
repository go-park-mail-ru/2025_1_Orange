package repository

type SessionRepository interface {
	CreateSession(userID int, role string) (string, error)
	GetSession(sessionToken string) (userID int, role string, err error)
	DeleteSession(sessionToken string) error
	DeleteAllSessions(userID int, role string) error
}
