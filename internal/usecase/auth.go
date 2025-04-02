package usecase

type Auth interface {
	Logout(session string) error
	LogoutAll(userID int, role string) error
	GetUserIDBySession(session string) (int, string, error)
	CreateSession(userID int, role string) (string, error)
}
