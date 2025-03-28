package mysql

import (
	"ResuMatch/internal/domain/mocks"
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type IUserRepository interface {
	AddUser(ctx context.Context, user *mocks.DBUser) error
	GetUserIdByEmail(ctx context.Context, email string) (int, error)
	GetUserInfo(ctx context.Context, userID int) (*mocks.DBUser, error)
	UpdateUserInfo(ctx context.Context, userID int, user *mocks.UserUpdateInfo) error
	checkPasswordById(ctx context.Context, UserId int, CheckPassword string) error
	castRawPasswordAndCompare(rawHash interface{}, passwordToCheck string) error
}

type UserRepository struct {
	userStorage *sql.DB
}

func NewUserRepository(db *sql.DB) IUserRepository {
	return &UserRepository{
		userStorage: db,
	}
}

func (p *UserRepository) castRawPasswordAndCompare(rawHash interface{}, passwordToCheck string) error {
	castedHash, ok := rawHash.([]byte)
	if !ok {
		return fmt.Errorf("The server encountered a problem and could not process your request")
	}

	isEqual := func(password string, hashedPass []byte) bool {
		hashToCheck := argon2.IDKey([]byte(password), 1, 64*1024, 4, 32)
		return subtle.ConstantTimeCompare(hashedPass, hashToCheck) == 1
	}(passwordToCheck, castedHash)

	if !isEqual {
		return fmt.Errorf("Incorrect credentials")
	}

	return nil
}

func (m *UserRepository) checkPasswordById(ctx context.Context, UserId int, CheckPassword string) error {
	var actualHash string

	query := `SELECT password FROM user WHERE id = ?`
	err := m.userStorage.QueryRow(query, UserId).Scan(&actualHash)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("user not found")
	} else if err != nil {
		return err
	}

	return m.castRawPasswordAndCompare(actualHash, CheckPassword) // Вызов функции сравнения пароля
}

func (m *UserRepository) GetUserIdByEmail(ctx context.Context, email string) (int, error) {
	var userID int

	query := `SELECT id FROM user WHERE email = ?`
	err := m.userStorage.QueryRow(query, email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, errors.New("user not found")
	} else if err != nil {
		return 0, err
	}

	return userID, nil
}
func (m *UserRepository) AddUser(ctx context.Context, user *mocks.DBUser) error {
	tx, err := m.userStorage.Begin()
	if err != nil {
		return err
	}
	var exists bool
	err = tx.QueryRow(`SELECT EXISTS (SELECT id FROM user WHERE email = ?)`, user.Email).Scan(&exists)
	if err != nil {
		tx.Rollback()
		return err
	}
	if exists {
		tx.Rollback()
		return errors.New("account already exists")
	}

	query := `INSERT INTO user 
        (email, password, first_name, last_name, birthday, city, gender, job_search_status_id, profile_image_id) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.Exec(query, user.Email, user.Password, user.FirstName, user.LastName, user.Birthday, user.City, user.Gender, user.JobSearchStatusId, user.ProfileImageId)
	if err != nil {
		tx.Rollback()
		return err
	}
	comerr := tx.Commit()
	if comerr != nil {
		return comerr
	}
	return nil
}

func (m *UserRepository) GetUserInfo(ctx context.Context, userID int) (*mocks.DBUser, error) {
	query := `SELECT id, email, first_name, last_name, birthday, city, gender, job_search_status_id, profile_image_id 
    FROM user WHERE id = ?`

	user := &mocks.DBUser{}
	err := m.userStorage.QueryRow(query, userID).Scan(&user.FirstName, &user.LastName, &user.Email, &user.Birthday, &user.City, &user.Gender, &user.JobSearchStatusId, &user.ProfileImageId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

func (m *UserRepository) UpdateUserInfo(ctx context.Context, userID int, user *mocks.UserUpdateInfo) error {
	query := `UPDATE user 
		SET email = ?, first_name = ?, last_name = ?, birthday = ?, city = ?, gender = ?, job_search_status_id = ?, profile_image_id = ? 
		WHERE id = ?`

	_, err := m.userStorage.Exec(query, user.Email, user.FirstName, user.LastName, user.Birthday, user.City, user.Gender, userID)
	if err != nil {
		return err
	}
	return nil
}
