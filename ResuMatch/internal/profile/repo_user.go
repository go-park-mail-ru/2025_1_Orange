package profile

import (
	"ResuMatch/internal/data"
	"ResuMatch/internal/models"
	"context"
	"errors"
	"fmt"
	"net/mail"
	//_ "github.com/jackc/pgx/stdlib"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrNotAllowed    = errors.New("not allowed")
	ErrInvalideEmail = errors.New("invalide email")
)

type UserStorage struct {
	Users map[string]models.User
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		Users: data.Users,
	}
}

func (u *UserStorage) CreateUser(email, password, firstname, lastname, companyname, companyaddress string) (models.User, error) {
	if _, exists := u.Users[email]; exists {
		return models.User{}, errors.New("email already exists")
	}
	newUser := models.User{
		ID:             uint64(len(u.Users) + 1),
		Email:          email,
		Password:       password,
		FirstName:      firstname,
		LastName:       lastname,
		CompanyName:    companyname,
		CompanyAddress: companyaddress,
	}
	u.Users[email] = newUser
	return newUser, nil
}

func (u *UserStorage) GetUserByEmail(email string) (*models.User, bool) {
	for _, user := range u.Users {
		if user.Email == email {
			return &user, true
		}
	}
	return nil, false
}

func (u *UserStorage) GetUserById(id uint64) (*models.User, bool) {
	for _, user := range u.Users {
		if user.ID == id {
			return &user, true
		}
	}
	return nil, false
}

func (u *UserStorage) CreateUserAccount(_ context.Context, email string, password string, firstname string, lastname string, companyname string, companyaddress string) (models.User, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return models.User{}, ErrInvalideEmail
	}
	user, err := u.CreateUser(email, password, firstname, lastname, companyname, companyaddress)
	if err != nil {
		return models.User{}, fmt.Errorf("CreateUserAccount err: %w", err)
	}

	return user, nil
}
