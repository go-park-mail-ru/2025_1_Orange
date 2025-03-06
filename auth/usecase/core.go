package usecase

import (
	"auth/models"
	"auth/repository/profile"
	"auth/repository/session"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sync"
	"time"
)

type ICore interface {
	CreateSession(ctx context.Context, login string) (string, session.Session, error)
	KillSession(ctx context.Context, sid string) error
	FindActiveSession(ctx context.Context, sid string) (bool, error)
	CreateUserAccount(login string, password string, name string, birthDate string, email string) error
	FindUserAccount(login string, password string) (*models.User, bool, error)
	FindUserByLogin(login string) (bool, error)
	GetUserName(ctx context.Context, sid string) (string, error)
	GetUserProfile(login string) (*models.User, error)
	GetUserRole(login string) (string, error)
	FindUsers(login string, role string, first, limit uint64) ([]models.User, error)
}

type Core struct {
	sessions session.SessionRepo
	mutex    sync.RWMutex
	users    profile.IUserRepo
}

var (
	ErrNotFound    = errors.New("not found")
	ErrNotAllowed  = errors.New("not allowed")
	LostConnection = errors.New("Redis connection lost")
	InvalideEmail  = errors.New("invalide email")
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func (core *Core) GetUserName(ctx context.Context, sid string) (string, error) {
	core.mutex.RLock()
	login, err := core.sessions.GetUserLogin(ctx, sid)
	core.mutex.RUnlock()

	if err != nil {
		return "", err
	}

	return login, nil
}

func (core *Core) CreateSession(ctx context.Context, login string) (string, session.Session, error) {
	sid := RandStringRunes(32)

	newSession := session.Session{
		Login:   login,
		SID:     sid,
		Expires: time.Now().Add(24 * time.Hour),
	}

	core.mutex.Lock()
	sessionAdded, err := core.sessions.AddSession(ctx, newSession)
	core.mutex.Unlock()

	if !sessionAdded && err != nil {
		return "", session.Session{}, err
	}

	if !sessionAdded {
		return "", session.Session{}, nil
	}

	return sid, newSession, nil
}

func (core *Core) FindActiveSession(ctx context.Context, sid string) (bool, error) {
	core.mutex.RLock()
	found, err := core.sessions.CheckActiveSession(ctx, sid)
	core.mutex.RUnlock()

	if err != nil {
		return false, err
	}

	return found, nil
}

func (core *Core) KillSession(ctx context.Context, sid string) error {
	core.mutex.Lock()
	_, err := core.sessions.DeleteSession(ctx, sid)
	core.mutex.Unlock()

	if err != nil {
		return err
	}

	return nil
}

func (core *Core) CreateUserAccount(login string, password string, name string, birthDate string, email string) error {
	if matched, _ := regexp.MatchString(`@`, email); !matched {
		return InvalideEmail
	}
	err := core.users.CreateUser(login, password, name, birthDate, email)
	if err != nil {
		return fmt.Errorf("CreateUserAccount err: %w", err)
	}

	return nil
}

func (core *Core) FindUserAccount(login string, password string) (*models.User, bool, error) {
	user, found, err := core.users.GetUser(login, password)
	if err != nil {
		return nil, false, fmt.Errorf("FindUserAccount err: %w", err)
	}
	return user, found, nil
}

func (core *Core) FindUserByLogin(login string) (bool, error) {
	found, err := core.users.FindUser(login)
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
	profile, err := core.users.GetUserProfile(login)
	if err != nil {
		return nil, fmt.Errorf("GetUserProfile err: %w", err)
	}

	return profile, nil
}

func (core *Core) GetUserRole(login string) (string, error) {
	role, err := core.users.GetUserRole(login)
	if err != nil {
		return "", fmt.Errorf("get user role err: %w", err)
	}

	return role, nil
}

func (core *Core) FindUsers(login string, role string, first, limit uint64) ([]models.User, error) {
	users, err := core.users.FindUsers(login, role, first, limit)
	if err != nil {
		return nil, fmt.Errorf("find user error: %w", err)
	}
	if len(users) == 0 {
		return nil, ErrNotFound
	}

	return users, nil
}
