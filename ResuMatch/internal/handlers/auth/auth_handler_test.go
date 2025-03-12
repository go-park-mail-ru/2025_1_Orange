package auth

import (
	"ResuMatch/internal/models"
	"ResuMatch/internal/profile"
	request "ResuMatch/internal/request"
	"ResuMatch/internal/session"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignup(t *testing.T) {

	handler := NewMyHandler()

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
	userRepo := &profile.UserStorage{
		Users: map[string]models.User{
			"user1": {
				ID:       1,
				Email:    "test@example.com",
				Password: "password123",
			},
		},
	}

	handler := NewMyHandler()
	handler.user = userRepo

	reqBody := request.SigninRequest{
		Email:    "test@example.com",
		Password: "password123",
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
	sessionRepo := &session.SessionStorage{
		Sessions: make(map[string]uint64),
	}
	sessionID := "valid-session-id"
	sessionRepo.Sessions[sessionID] = 1

	handler := NewMyHandler()

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
	// Подготовка тестовых данных
	userRepo := &profile.UserStorage{
		Users: map[string]models.User{
			"user1": {
				ID:    1,
				Email: "user1@example.com",
			},
		},
	}
	handler := NewMyHandler()
	handler.user = userRepo

	// Тест 1: Проверка существующего email
	existingEmail := userRepo.Users["user1"].Email
	requestBody := map[string]string{"email": existingEmail}
	jsonBody, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/check-email", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CheckEmail(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Test Case 1 Failed: Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	var response map[string]string
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	expectedMessage := "Email already exists"
	if response["message"] != expectedMessage {
		t.Errorf("Test Case 1 Failed: Expected message %s, got %s", expectedMessage, response["message"])
	}

	// Тест 2: Проверка несуществующего email
	nonExistingEmail := "nonexisting@example.com"
	requestBody = map[string]string{"email": nonExistingEmail}
	jsonBody, _ = json.Marshal(requestBody)

	req = httptest.NewRequest(http.MethodPost, "/check-email", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	handler.CheckEmail(w, req)

	res = w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Test Case 2 Failed: Expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response body: %v", err)
	}

	expectedMessage = "Email not found"
	if response["message"] != expectedMessage {
		t.Errorf("Test Case 2 Failed: Expected message %s, got %s", expectedMessage, response["message"])
	}
}
func TestAuth(t *testing.T) {
	userRepo := &profile.UserStorage{
		Users: map[string]models.User{
			"user1": {
				ID:    1,
				Email: "user1@example.com",
			},
		},
	}

	sessionRepo := &session.SessionStorage{
		Sessions: make(map[string]uint64),
	}

	handler := NewMyHandler()
	handler.user = userRepo
	handler.session = sessionRepo

	existingUser := userRepo.Users["user1"]
	sessionID, err := sessionRepo.CreateSession(context.Background(), existingUser.ID)
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

	handler := NewMyHandler()

	invalidBody := `{"email": "test@example.com", "password": }`
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

	handler := NewMyHandler()

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
	userRepo := &profile.UserStorage{
		Users: map[string]models.User{
			"user1": {
				ID:       1,
				Email:    "user1@example.com",
				Password: "correctpassword",
			},
		},
	}

	handler := NewMyHandler()
	handler.user = userRepo

	reqBody := request.SigninRequest{
		Email:    "user1@example.com",
		Password: "wrongpassword", // Неверный пароль
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

	handler := NewMyHandler()

	reqBody := request.SignupRequest{
		Email:          "user@example.com",
		Password:       "password123",
		RepeatPassword: "password321",
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
	handler := NewMyHandler()

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
	handler := NewMyHandler()

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
	handler := NewMyHandler()

	// Создаем запрос с некорректным телом (не JSON)
	invalidBody := "invalid json"
	req := httptest.NewRequest(http.MethodPost, "/check-email", bytes.NewBufferString(invalidBody))
	req.Header.Set("Content-Type", "application/json")
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

	expectedMessage := "Invalid request body"
	if response["error"] != expectedMessage {
		t.Errorf("Expected error message %s, got %s", expectedMessage, response["error"])
	}
}
func TestAuth_NoSessionCookie(t *testing.T) {

	handler := NewMyHandler()

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

	handler := NewMyHandler()

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
