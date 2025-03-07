package usecase

import (
	"ResuMatch/models"
	"ResuMatch/repository/profile"
	"ResuMatch/repository/session"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/mail"
)

type ICore interface {
	CreateSession(ctx context.Context, userID uint64) (string, error)
	FindActiveSession(ctx context.Context, sid string) (uint64, error)
	KillSession(ctx context.Context, sid string) error
	CreateUserAccount(login string, password string, name string, birthDate string, email string) error
	FindUserAccount(login string, password string) (*models.User, bool, error)
	FindUserByLogin(login string) (bool, error)
	GetUserName(ctx context.Context, sid string) (string, error)
	GetUserProfile(login string) (*models.User, error)
	GetUserRole(login string) (string, error)
	FindUsers(login string, role string, first, limit uint64) ([]models.User, error)
}

type Core struct {
	Sessions session.SessionRepo
	Users    profile.UserRepo
}

func NewCore(sessions session.SessionRepo, users profile.UserRepo) *Core {
	return &Core{
		Sessions: sessions,
		Users:    users,
	}
}

var (
	ErrNotFound   = errors.New("not found")
	ErrNotAllowed = errors.New("not allowed")
	InvalideEmail = errors.New("invalide email")
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func CreateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("createSessionID: failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (core *Core) CreateSession(ctx context.Context, userID uint64) (string, error) {
	sid, err := CreateSessionID()
	if err != nil {
		return "", fmt.Errorf("CreateSession: can't generate session ID for user %d: %w", userID, err)
	}

	err = core.Sessions.CreateSession(ctx, userID, sid)
	if err != nil {
		return "", fmt.Errorf("CreateSession: can't create session for user %d with sid %s: %w", userID, sid, err)
	}
	return sid, nil
}
func (core *Core) FindActiveSession(ctx context.Context, sid string) (uint64, error) {
	userID, err := core.Sessions.GetSession(sid)
	if err != nil {
		return 0, fmt.Errorf("FindActiveSession: can't get session %s: %w", sid, err)
	}
	return userID, nil
}

func (core *Core) KillSession(ctx context.Context, sid string) error {
	err := core.Sessions.DeleteSession(sid)
	if err != nil {
		return fmt.Errorf("KillSession: can't delete session %s: %w", sid, err)
	}
	return nil
}

func (core *Core) GetUserName(ctx context.Context, sid string) (string, error) {
	userID, err := core.Sessions.GetSession(sid)
	if err != nil {
		return "", fmt.Errorf("GetUserName: can't get userID for session %s: %w", sid, err)
	}
	login, err := core.Users.GetLoginByID(userID)
	if err != nil {
		return "", fmt.Errorf("GetUserName: can't get username for userID %d: %w", userID, err)
	}

	return login, nil
}

func (core *Core) CreateUserAccount(c context.Context, login string, password string, name string, birthDate string, email string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return InvalideEmail
	}
	err := core.Users.CreateUser(login, password, name, birthDate, email)
	if err != nil {
		return fmt.Errorf("CreateUserAccount err: %w", err)
	}

	return nil
}

func (core *Core) FindUserAccount(login string, password string) (*models.User, bool, error) {
	user, found, err := core.Users.GetUser(login, password)
	if err != nil {
		return nil, false, fmt.Errorf("FindUserAccount err: %w", err)
	}
	return user, found, nil
}

func (core *Core) FindUserByLogin(login string) (bool, error) {
	found, err := core.Users.FindUser(login)
	if err != nil {
		return false, fmt.Errorf("FindUserByLogin err: %w", err)
	}

	return found, nil
}

func RandStringRunes(seed int) string {
	symbols := make([]rune, seed)
	for i := range symbols {
		symbols[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(symbols)
}

func (core *Core) GetUserProfile(login string) (*models.User, error) {
	profile, err := core.Users.GetUserProfile(login)
	if err != nil {
		return nil, fmt.Errorf("GetUserProfile err: %w", err)
	}

	return profile, nil
}

func (core *Core) GetUserRole(login string) (string, error) {
	role, err := core.Users.GetUserRole(login)
	if err != nil {
		return "", fmt.Errorf("get user role err: %w", err)
	}

	return role, nil
}

func (core *Core) FindUsers(login string, role string, first, limit uint64) ([]models.User, error) {
	users, err := core.Users.FindUsers(login, role, first, limit)
	if err != nil {
		return nil, fmt.Errorf("find user error: %w", err)
	}
	if len(users) == 0 {
		return nil, ErrNotFound
	}

	return users, nil
}
