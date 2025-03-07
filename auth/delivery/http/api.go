package main

import (
	request "auth/request"
	"auth/usecase"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"
)

type MyHandler struct {
	core *usecase.Core
}

func NewMyHandler(core *usecase.Core) *MyHandler {
	return &MyHandler{
		core: core,
	}
}

func (api *MyHandler) Signin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req request.SigninRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		log.Println(err)
		return
	}

	user, ok := api.core.Users.GetUserByLogin(req.Login)
	if !ok {
		http.Error(w, `{"error": "no user"}`, http.StatusNotFound)
		return
	}

	if user.Password != req.Password {
		http.Error(w, `{"error": "bad pass"}`, http.StatusUnauthorized)
		return
	}
	sid, err := api.core.CreateSession(r.Context(), user.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to create session: %s"}`, err.Error()), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sid,
		Expires:  time.Now().Add(10 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)

	json.NewEncoder(w).Encode(map[string]string{"session_id": sid})
}

func (api *MyHandler) Signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req request.SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		log.Println(err)
		return
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		http.Error(w, `{"error": "Invalid email format"}`, http.StatusBadRequest)
		return
	}

	_, exists := api.core.Users.GetUserByLogin(req.Login)
	if exists {
		http.Error(w, `{"error": "Login already exists"}`, http.StatusConflict)
		return
	}

	err := api.core.CreateUserAccount(r.Context(), req.Login, req.Password, req.Name, req.BirthDate, req.Email)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to create user: %s"}`, err.Error()), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}
func (api *MyHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sessionCookie, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Error(w, `{"error": "no session"}`, http.StatusUnauthorized)
		return
	}

	sid := sessionCookie.Value

	err = api.core.KillSession(r.Context(), sid)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "failed to kill session: %s"}`, err.Error()), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	sessionCookie.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, sessionCookie)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
