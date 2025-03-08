package auth

import (
	"ResuMatch/internal/models"
	"ResuMatch/internal/repository/profile"
	"ResuMatch/internal/repository/session"
	request "ResuMatch/internal/request"
	"ResuMatch/internal/usecase"
	"bytes"

	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Глобальные переменные для хранения данных
var Sessions = make(map[string]uint64)   // Реальная структура для хранения сессий
var Users = make(map[string]models.User) // Реальная структура для хранения пользователей

// Реализация SessionRepo без mock
type RealSessionRepo struct{}

func (r *RealSessionRepo) CreateSession(_ context.Context, userID uint64, sid string) error {
	Sessions[sid] = userID
	return nil
}

func (r *RealSessionRepo) GetSession(sessionID string) (uint64, error) {
	userID, ok := Sessions[sessionID]
	if !ok {
		return 0, fmt.Errorf("session not found")
	}
	return userID, nil
}

func (r *RealSessionRepo) DeleteSession(sessionID string) error {
	delete(Sessions, sessionID)
	return nil
}

// Реализация UserRepo без mock
type RealUserRepo struct{}

func (r *RealUserRepo) GetUserByEmail(email string) (*models.User, bool) {
	user, ok := Users[email]
	if !ok {
		return nil, false
	}
	return &user, true
}

func (r *RealUserRepo) CreateUser(email, password, firstname, lastname, companyname, companyaddress string) error {
	if _, exists := Users[email]; exists {
		return fmt.Errorf("email already exists")
	}
	Users[email] = models.User{
		ID:             uint64(len(Users) + 1),
		Email:          email,
		Password:       password,
		FirstName:      firstname,
		LastName:       lastname,
		CompanyName:    companyname,
		CompanyAddress: companyaddress,
	}
	return nil
}

// Тест для обработчика Signin
func TestSignin(t *testing.T) {
	// Настройка реальных данных
	sessionRepo := session.Sessionrepo{}
	userRepo := profile.UserRepo{}
	Users["test@example.com"] = models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: "password123",
	}

	// Создание Core и обработчика
	core := usecase.NewCore(sessionRepo, userRepo)
	handler := NewMyHandler(core)

	// Создание запроса
	reqBody := request.SigninRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Вызов обработчика
	handler.Signin(w, req)

	// Проверка ответа
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
}

// Тест для обработчика Signup
func TestSignup(t *testing.T) {
	// Настройка реальных данных
	sessionRepo := session.Sessionrepo{}
	userRepo := profile.UserRepo{}

	// Создание Core и обработчика
	core := usecase.NewCore(sessionRepo, userRepo)
	handler := NewMyHandler(core)

	// Создание запроса
	reqBody := request.SignupRequest{
		Email:          "newuser@example.com",
		Password:       "password123",
		RepeatPassword: "password123",
		FirstName:      "John",
		LastName:       "Doe",
		CompanyName:    "Test Co",
		CompanyAddress: "123 Test Street",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Вызов обработчика
	handler.Signup(w, req)

	// Проверка ответа
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}
}

// Тест для обработчика Logout
func TestLogout(t *testing.T) {
	// Настройка реальных данных
	sessionRepo := session.Sessionrepo{}
	userRepo := profile.UserRepo{}
	Sessions["valid-session-id"] = 1

	// Создание Core и обработчика
	core := usecase.NewCore(sessionRepo, userRepo)
	handler := NewMyHandler(core)

	// Создание запроса
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session-id"})
	w := httptest.NewRecorder()

	// Вызов обработчика
	handler.Logout(w, req)

	// Проверка ответа
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
}
