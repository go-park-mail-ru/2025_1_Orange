package auth

import (
	"ResuMatch/internal/data"
	"ResuMatch/internal/models"
	"ResuMatch/internal/repository/profile"
	"ResuMatch/internal/repository/session"
	request "ResuMatch/internal/request"
	"ResuMatch/internal/usecase"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignup(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}

	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	reqBody := request.SignupRequest{
		Email:          "newuser@example.com",
		Password:       "password123",
		RepeatPassword: "password123",
		FirstName:      "New",
		LastName:       "User",
		CompanyName:    "New Co",
		CompanyAddress: "123 Test Street",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Signup(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}
}
func TestSignin(t *testing.T) {

	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}

	existingUser := data.Users["user1"]
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	reqBody := request.SigninRequest{
		Email:    existingUser.Email,
		Password: existingUser.Password,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Signin(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
}

func TestLogout(t *testing.T) {

	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}

	sessionID := "valid-session-id"
	session.Sessions[sessionID] = 1

	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}
}

func TestCheckEmail(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}

	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	existingEmail := data.Users["user1"].Email

	// Тест на существующий email
	req := httptest.NewRequest(http.MethodGet, "/check-email?email="+existingEmail, nil)
	w := httptest.NewRecorder()

	handler.CheckEmail(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Test Case 1 Failed: Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	// Тест на несуществующий email
	nonExistingEmail := "nonexisting@example.com"
	req = httptest.NewRequest(http.MethodGet, "/check-email?email="+nonExistingEmail, nil)
	w = httptest.NewRecorder()

	handler.CheckEmail(w, req)

	res = w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Test Case 2 Failed: Expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}
func TestAuth(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}

	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	existingUser := data.Users["user1"]
	sessionID, err := core.CreateSession(context.Background(), existingUser.ID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	}
	req.AddCookie(cookie)

	w := httptest.NewRecorder()
	handler.Auth(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var responseUser models.User
	if err := json.NewDecoder(res.Body).Decode(&responseUser); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if responseUser.ID != existingUser.ID {
		t.Errorf("Expected user ID %d, got %d", existingUser.ID, responseUser.ID)
	}
}

func TestSignin_InvalidJSON(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	invalidBody := `{"email": "test@example.com", "password": }` // Неверный JSON
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader([]byte(invalidBody)))
	w := httptest.NewRecorder()

	handler.Signin(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}

func TestSignin_UserNotFound(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	reqBody := request.SigninRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Signin(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestSignin_WrongPassword(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	existingUser := data.Users["user1"]

	reqBody := request.SigninRequest{
		Email:    existingUser.Email,
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Signin(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, res.StatusCode)
	}
}

func TestSignup_PasswordMismatch(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	reqBody := request.SignupRequest{
		Email:          "user@example.com",
		Password:       "password123",
		RepeatPassword: "password321", // Не совпадают
		FirstName:      "Test",
		LastName:       "User",
		CompanyName:    "TestCo",
		CompanyAddress: "123 Test St",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Signup(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}

func TestSignup_InvalidEmail(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	reqBody := request.SignupRequest{
		Email:          "invalid-email",
		Password:       "password123",
		RepeatPassword: "password123",
		FirstName:      "Test",
		LastName:       "User",
		CompanyName:    "TestCo",
		CompanyAddress: "123 Test St",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Signup(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}

func TestLogout_NoSessionCookie(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}

	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	// Нет cookie сессии
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, res.StatusCode)
	}
}

func TestCheckEmail_InvalidRequestBody(t *testing.T) {
	// Подготовка необходимых объектов
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	// Пустой URL без email-параметра (CheckEmail ожидает email в query params)
	req := httptest.NewRequest(http.MethodGet, "/check-email", nil)
	w := httptest.NewRecorder()

	// Вызов метода CheckEmail
	handler.CheckEmail(w, req)

	// Проверка ответа
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}

	// Проверка содержимого ответа
	var response map[string]string
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	expectedMessage := "Email parameter is required"
	if response["error"] != expectedMessage {
		t.Errorf("Expected error message %s, got %s", expectedMessage, response["error"])
	}
}
func TestAuth_NoSessionCookie(t *testing.T) {
	sessionRepo := &session.Sessionrepo{}
	userRepo := &profile.UserRepo{}
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	// Создаем тестовый запрос без cookie
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	w := httptest.NewRecorder()

	// Вызываем функцию Auth
	handler.Auth(w, req)

	// Проверяем статус-код на Unauthorized (401)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, res.StatusCode)
	}

	// Проверяем, что в ответе присутствует сообщение об ошибке
	var response map[string]string
	err := json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Fatal("Failed to decode response body:", err)
	}

	if response["error"] != "Unauthorized" {
		t.Errorf("Expected error 'Unauthorized', got %v", response["error"])
	}
}

func TestAuth_InvalidSession(t *testing.T) {
	userRepo := &profile.UserRepo{}
	sessionRepo := &session.Sessionrepo{}
	core := usecase.NewCore(*sessionRepo, *userRepo)
	handler := NewMyHandler(core)

	// Создаем запрос с cookie, но с невалидным значением session_id
	req := httptest.NewRequest(http.MethodGet, "/auth", nil)
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: "invalid_session_id",
	}
	req.AddCookie(cookie)
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.Auth(w, req)

	// Проверяем статус код - должен быть Unauthorized (401)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, res.StatusCode)
	}
}
