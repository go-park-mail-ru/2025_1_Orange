package auth

import (
	"ResuMatch/internal/data"
	"ResuMatch/internal/repository/profile"
	"ResuMatch/internal/repository/session"
	request "ResuMatch/internal/request"
	"ResuMatch/internal/usecase"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
