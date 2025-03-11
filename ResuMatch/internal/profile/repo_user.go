package profile

import (
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

type IUserRepo interface {
	GetUserByEmail(email string) (*models.User, bool)
	GetUserById(Id uint64) (*models.User, bool)
	CreateUser(email string, password string, name string, birthDate string) error
}

type UserStorage struct {
	Users map[string]models.User
}

func NewUserStorage() *UserStorage {
	return &UserStorage{
		Users: make(map[string]models.User),
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

// type UserRepo struct{}

// func (r UserRepo) GetUserById(Id uint64) (*models.User, bool) {
// 	for i := range data.Users {
// 		if data.Users[i].ID == Id {
// 			user := data.Users[i]
// 			return &user, true
// 		}
// 	}
// 	return nil, false
// }

// func (r UserRepo) GetUserByEmail(email string) (*models.User, bool) {
// 	for i := range data.Users {
// 		if data.Users[i].Email == email {
// 			user := data.Users[i]
// 			return &user, true
// 		}
// 	}
// 	return nil, false
// }

// func (r UserRepo) CreateUser(email string, password string, firstname string, lastname string, companyname string, companyaddress string) (models.User, error) {
// 	for _, user := range data.Users {
// 		if user.Email == email {
// 			return models.User{}, errors.New("email already exists")
// 		}
// 	}
// 	newUser := models.User{
// 		ID:             uint64(len(data.Users) + 1),
// 		Email:          email,
// 		Password:       password,
// 		FirstName:      firstname,
// 		LastName:       lastname,
// 		CompanyName:    companyname,
// 		CompanyAddress: companyaddress,
// 	}

// 	data.Users[email] = newUser

// 	return newUser, nil
// }

// func (r UserRepo) GetUser(email string, password string) (*models.User, bool, error) {
// 	for i := range data.Users {
// 		if data.Users[i].Email == email && data.Users[i].Password == password {
// 			user := data.Users[i]
// 			return &user, true, nil
// 		}
// 	}
// 	return nil, false, fmt.Errorf("GetUser err")
// }

// func (r UserRepo) FindUser(email string) (bool, error) {
// 	for i := range data.Users {
// 		if data.Users[i].Email == email {
// 			return true, nil
// 		}
// 	}
// 	return false, fmt.Errorf("GetUserProfileId err")
// }

// func (r UserRepo) FindUsers(email string, _ string, first, limit uint64) ([]models.User, error) {
// 	var foundUsers []models.User
// 	count := uint64(0)

// 	for _, user := range data.Users {
// 		if email != "" && !strings.Contains(user.Email, email) {
// 			continue
// 		}
// 		if count >= first {
// 			foundUsers = append(foundUsers, user)
// 		}
// 		count++
// 		if len(foundUsers) >= int(limit) {
// 			break
// 		}
// 	}
// 	if len(foundUsers) == 0 {
// 		return nil, fmt.Errorf("Users not found")
// 	}

//		return foundUsers, nil
//	}

// func (r UserRepo) GetUserProfile(email string) (*models.User, error) {
// 	for i := range data.Users {
// 		if data.Users[i].Email == email {
// 			user := data.Users[i]
// 			return &user, nil
// 		}
// 	}
// 	return nil, fmt.Errorf("GetUserProfileId err")
// }

// func (r UserRepo) GetUserCompany(email string) (string, error) {
// 	for i := range data.Users {
// 		if data.Users[i].Email == email {
// 			return data.Users[i].CompanyName, nil
// 		}
// 	}
// 	return "", fmt.Errorf("GetUserRole err")
// }

// func (r UserRepo) GetEmailByID(userID uint64) (string, error) {
// 	for email, user := range data.Users {
// 		if user.ID == userID {
// 			return email, nil
// 		}
// 	}
// 	return "", fmt.Errorf("user with ID %d not found", userID)
// }
