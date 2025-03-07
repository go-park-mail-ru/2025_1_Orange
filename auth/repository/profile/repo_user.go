package profile

import (
	"auth/data"
	"auth/models"
	"errors"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/stdlib"
)

type IUserRepo interface {
	GetUser(login string, password string) (*models.User, bool, error)
	GetUserProfileId(login string) (int64, error)
	FindUser(login string) (bool, error)
	CreateUser(login string, password string, name string, birthDate string, email string) error
	GetUserProfile(login string) (*models.User, error)
	CheckUserPassword(login string, password string) (bool, error)
	GetUserRole(login string) (string, error)
	FindUsers(login string, role string, first, limit uint64) ([]models.User, error)
	GetLoginByID(userID uint64) (string, error)
	GetUserByLogin(login string) (*models.User, bool)
}

type UserRepo struct{}

func (r UserRepo) GetUserByLogin(login string) (*models.User, bool) {
	user, ok := data.Users[login]
	if !ok {
		return nil, false
	}
	return &user, true
}
func (r UserRepo) GetUser(login string, password string) (*models.User, bool, error) {
	for i := range data.Users {
		if data.Users[i].Login == login && data.Users[i].Password == password {
			user := data.Users[i]
			return &user, true, nil
		}
	}
	return nil, false, fmt.Errorf("GetUser err")
}

func (r UserRepo) FindUser(login string) (bool, error) {
	for i := range data.Users {
		if data.Users[i].Login == login {
			return true, nil
		}
	}
	return false, fmt.Errorf("GetUserProfileId err")
}

func (r UserRepo) FindUsers(login string, role string, first, limit uint64) ([]models.User, error) {
	var foundUsers []models.User
	count := uint64(0)

	for _, user := range data.Users {
		if login != "" && !strings.Contains(user.Login, login) {
			continue
		}
		if count >= first {
			foundUsers = append(foundUsers, user)
		}
		count++
		if len(foundUsers) >= int(limit) {
			break
		}
	}
	if len(foundUsers) == 0 {
		return nil, fmt.Errorf("Users not found")
	}

	return foundUsers, nil
}
func (r UserRepo) CreateUser(login string, password string, name string, birthDate string, email string) error {
	for _, user := range data.Users {
		if user.Login == login {
			return errors.New("login already exists")
		}
	}
	newUser := models.User{
		Id:               uint64(len(data.Users) + 1),
		Name:             name,
		Login:            login,
		Password:         password,
		Birthdate:        birthDate,
		Photo:            "",
		RegistrationDate: "2023-11-03",
		Email:            email,
		Role:             "user",
	}

	data.Users[login] = newUser

	return nil
}

func (r UserRepo) GetUserProfile(login string) (*models.User, error) {
	for i := range data.Users {
		if data.Users[i].Login == login {
			user := data.Users[i]
			return &user, nil
		}
	}
	return nil, fmt.Errorf("GetUserProfileId err")
}

func (r UserRepo) GetUserRole(login string) (string, error) {
	for i := range data.Users {
		if data.Users[i].Login == login {
			return data.Users[i].Role, nil
		}
	}
	return "", fmt.Errorf("GetUserRole err")
}

func (r UserRepo) GetLoginByID(userID uint64) (string, error) {
	for login, user := range data.Users {
		if user.Id == userID {
			return login, nil
		}
	}
	return "", fmt.Errorf("user with ID %d not found", userID)
}
