package profile

import (
	"ResuMatch/internal/data"
	"ResuMatch/internal/models"
	"errors"
	"fmt"
	"strings"
	//_ "github.com/jackc/pgx/stdlib"
)

type IUserRepo interface {
	GetUser(email string, password string) (*models.User, bool, error)
	GetUserProfileId(email string) (int64, error)
	FindUser(email string) (bool, error)
	CreateUser(email string, password string, name string, birthDate string) error
	GetUserProfile(email string) (*models.User, error)
	CheckUserPassword(email string, password string) (bool, error)
	GetUserCompany(email string) (string, error)
	FindUsers(email string, role string, first, limit uint64) ([]models.User, error)
	GetEmailByID(userID uint64) (string, error)
	GetUserByEmail(email string) (*models.User, bool)
}

type UserRepo struct{}

func (r UserRepo) GetUserByEmail(email string) (*models.User, bool) {
	user, ok := data.Users[email]
	if !ok {
		return nil, false
	}
	return &user, true
}
func (r UserRepo) GetUser(email string, password string) (*models.User, bool, error) {
	for i := range data.Users {
		if data.Users[i].Email == email && data.Users[i].Password == password {
			user := data.Users[i]
			return &user, true, nil
		}
	}
	return nil, false, fmt.Errorf("GetUser err")
}

func (r UserRepo) FindUser(email string) (bool, error) {
	for i := range data.Users {
		if data.Users[i].Email == email {
			return true, nil
		}
	}
	return false, fmt.Errorf("GetUserProfileId err")
}

func (r UserRepo) FindUsers(email string, _ string, first, limit uint64) ([]models.User, error) {
	var foundUsers []models.User
	count := uint64(0)

	for _, user := range data.Users {
		if email != "" && !strings.Contains(user.Email, email) {
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
func (r UserRepo) CreateUser(email string, password string, firstname string, lastname string, companyname string, companyaddress string) error {
	for _, user := range data.Users {
		if user.Email == email {
			return errors.New("email already exists")
		}
	}
	newUser := models.User{
		ID:             uint64(len(data.Users) + 1),
		Email:          email,
		Password:       password,
		FirstName:      firstname,
		LastName:       lastname,
		CompanyName:    companyname,
		CompanyAddress: companyaddress,
	}

	data.Users[email] = newUser

	return nil
}

func (r UserRepo) GetUserProfile(email string) (*models.User, error) {
	for i := range data.Users {
		if data.Users[i].Email == email {
			user := data.Users[i]
			return &user, nil
		}
	}
	return nil, fmt.Errorf("GetUserProfileId err")
}

func (r UserRepo) GetUserCompany(email string) (string, error) {
	for i := range data.Users {
		if data.Users[i].Email == email {
			return data.Users[i].CompanyName, nil
		}
	}
	return "", fmt.Errorf("GetUserRole err")
}

func (r UserRepo) GetEmailByID(userID uint64) (string, error) {
	for email, user := range data.Users {
		if user.ID == userID {
			return email, nil
		}
	}
	return "", fmt.Errorf("user with ID %d not found", userID)
}
