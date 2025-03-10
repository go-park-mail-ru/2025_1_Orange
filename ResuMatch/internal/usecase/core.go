package usecase

import (
	"ResuMatch/internal/models"
	"ResuMatch/internal/repository/profile"
	"ResuMatch/internal/repository/session"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
)

type ICore interface {
	CreateSession(ctx context.Context, userID uint64) (string, error)
	FindActiveSession(ctx context.Context, sid string) (uint64, error)
	KillSession(ctx context.Context, sid string) error
	GetUserIDFromSession(sid string) (uint64, error)
	CreateUserAccount(email string, password string, firstname string, lastname string, companyname string, companyaddress string) error
	// FindUserAccount(email string, password string) (*models.User, bool, error)
	// FindUserByEmail(email string) (bool, error)
	// GetUserName(ctx context.Context, sid string) (string, error)
	// GetUserProfile(email string) (*models.User, error)
	// GetUserCompany(email string) (string, error)
	// FindUsers(email string, role string, first, limit uint64) ([]models.User, error)
}

type Core struct {
	Sessions session.Sessionrepo
	Users    profile.UserRepo
}

func NewCore(sessions session.Sessionrepo, users profile.UserRepo) *Core {
	return &Core{
		Sessions: sessions,
		Users:    users,
	}
}

var (
	ErrNotFound      = errors.New("not found")
	ErrNotAllowed    = errors.New("not allowed")
	ErrInvalideEmail = errors.New("invalide email")
)

//var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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
func (core *Core) FindActiveSession(_ context.Context, sid string) (uint64, error) {
	userID, err := core.Sessions.GetSession(sid)
	if err != nil {
		return 0, fmt.Errorf("FindActiveSession: can't get session %s: %w", sid, err)
	}
	return userID, nil
}

func (core *Core) KillSession(_ context.Context, sid string) error {
	err := core.Sessions.DeleteSession(sid)
	if err != nil {
		return fmt.Errorf("KillSession: can't delete session %s: %w", sid, err)
	}
	return nil
}

func (core *Core) GetUserIDFromSession(sid string) (uint64, error) {
	userID, err := core.Sessions.GetSession(sid)
	if err != nil {
		return 0, fmt.Errorf("GetUserIDFromSession: can't get session %s: %w", sid, err)
	}

	return userID, nil
}

func (core *Core) CreateUserAccount(_ context.Context, email string, password string, firstname string, lastname string, companyname string, companyaddress string) (models.User, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return models.User{}, ErrInvalideEmail
	}
	user, err := core.Users.CreateUser(email, password, firstname, lastname, companyname, companyaddress)
	if err != nil {
		return models.User{}, fmt.Errorf("CreateUserAccount err: %w", err)
	}

	return user, nil
}

// func (core *Core) GetUserName(_ context.Context, sid string) (string, error) {
// 	userID, err := core.Sessions.GetSession(sid)
// 	if err != nil {
// 		return "", fmt.Errorf("GetUserName: can't get userID for session %s: %w", sid, err)
// 	}
// 	email, err := core.Users.GetEmailByID(userID)
// 	if err != nil {
// 		return "", fmt.Errorf("GetUserName: can't get username for userID %d: %w", userID, err)
// 	}

// 	return email, nil
// }

// func (core *Core) FindUserAccount(email string, password string) (*models.User, bool, error) {
// 	user, found, err := core.Users.GetUser(email, password)
// 	if err != nil {
// 		return nil, false, fmt.Errorf("FindUserAccount err: %w", err)
// 	}
// 	return user, found, nil
// }

// func (core *Core) FindUserByEmail(email string) (bool, error) {
// 	found, err := core.Users.FindUser(email)
// 	if err != nil {
// 		return false, fmt.Errorf("FindUserByEmail err: %w", err)
// 	}

// 	return found, nil
// }

// func RandStringRunes(seed int) string {
// 	symbols := make([]rune, seed)
// 	for i := range symbols {
// 		randomIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
// 		symbols[i] = letterRunes[randomIndex.Int64()]
// 	}
// 	return string(symbols)
// }

// func (core *Core) GetUserProfile(email string) (*models.User, error) {
// 	profile, err := core.Users.GetUserProfile(email)
// 	if err != nil {
// 		return nil, fmt.Errorf("GetUserProfile err: %w", err)
// 	}

// 	return profile, nil
// }

// func (core *Core) GetUserCompany(email string) (string, error) {
// 	role, err := core.Users.GetUserCompany(email)
// 	if err != nil {
// 		return "", fmt.Errorf("get user role err: %w", err)
// 	}

// 	return role, nil
// }

// func (core *Core) FindUsers(email string, role string, first, limit uint64) ([]models.User, error) {
// 	users, err := core.Users.FindUsers(email, role, first, limit)
// 	if err != nil {
// 		return nil, fmt.Errorf("find user error: %w", err)
// 	}
// 	if len(users) == 0 {
// 		return nil, ErrNotFound
// 	}

// 	return users, nil
// }
