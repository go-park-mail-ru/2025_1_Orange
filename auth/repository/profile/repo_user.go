package profile

import (
	"auth/data"
	"auth/models"
	"database/sql"
	"errors"
	"fmt"

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
}

type RepoPostgre struct {
	db *sql.DB
}

func GetUser(login string, password string) (*models.User, bool, error) {
	for i := range data.Users {
		if data.Users[i].Login == login && data.Users[i].Password == password {
			return &data.Users[i], true, nil
		}
	}
	return nil, false, fmt.Errorf("GetUser err")
}

func FindUser(login string) (bool, error) {
	for i := range data.Users {
		if data.Users[i].Login == login {
			return true, nil
		}
	}
	return false, fmt.Errorf("GetUserProfileId err")
}

func GetUserProfileId(login string) (int64, error) {
	for i := range data.Users {
		if data.Users[i].Login == login {
			return &data.Users[i].Id, nil
		}
	}
	return 0, fmt.Errorf("GetUserProfileId err")
}

func CreateUser(login string, password string, name string, birthDate string, email string) error {
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

	data.Users = append(data.Users, newUser)

	return nil
}

func GetUserProfile(login string) (*models.User, error) {
	for i := range data.Users {
		if data.Users[i].Login == login {
			return data.Users[i], nil
		}
	}
	return nil, fmt.Errorf("GetUserProfileId err")
}

func GetUserRole(login string) (string, error) {
	for i := range data.Users {
		if data.Users[i].Login == login {
			return Users[i].Role, nil
		}
	}
	return "", fmt.Errorf("GetUserRole err")
}
