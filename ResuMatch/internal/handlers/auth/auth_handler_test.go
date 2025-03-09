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

	reqBody := request.CheckUserRequest{
		Email: existingEmail,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/check-email", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CheckEmail(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Test Case 1 Failed: Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	nonExistingEmail := "nonexisting@example.com"
	reqBody = request.CheckUserRequest{
		Email: nonExistingEmail,
	}
	body, _ = json.Marshal(reqBody)
	req = httptest.NewRequest(http.MethodPost, "/check-email", bytes.NewReader(body))
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
