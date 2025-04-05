package usecase

type Auth interface {
	Logout(string) error
	LogoutAll(int, string) error
	GetUserIDBySession(string) (int, string, error)
	CreateSession(int, string) (string, error)
}
