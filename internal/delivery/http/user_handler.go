package http

import (
	mocks "ResuMatch/internal/domain/mocks"
	usecase "ResuMatch/internal/usecase"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	uUsecase usecase.AuthUsecase
}

func NewUserHandler(r mux.Router, u usecase.AuthUsecase) {
	handler := &UserHandler{
		uUsecase: u,
	}
	router := mux.NewRouter()

	router.HandleFunc("/login", handler.Signin)
	router.HandleFunc("/signup", handler.Signup)
	router.HandleFunc("/logout", handler.Logout)
}

func (u UserHandler) Signin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	user := new(mocks.User)
	if err := json.NewDecoder(r.Body).Decode(&r); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		log.Println(err)
		return
	}

	//TODO сделать ошибку в логгер
	expiryTime := time.Now().Add(10 * time.Hour)

	sessionID, loginErr := u.uUsecase.Login(r.Context(), user.Email, user.Password, expiryTime.Unix())
	if loginErr != nil {
		//TODO сделать ошибку в логгер
		return
	}

	cookie := &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Expires:  expiryTime,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

func (u UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	user := new(mocks.DBUser)
	if err := json.NewDecoder(r.Body).Decode(&r); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		log.Println(err)
		return
	}
	expiryTime := time.Now().Add(10 * time.Hour)

	sessionID, err := u.uUsecase.SignUp(r.Context(), user, expiryTime.Unix())
	if err != nil {
		//TODO сделать ошибку в логгер
		return
	}

	cookie := &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Expires:  expiryTime,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}

func (u UserHandler) Logout(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session_id")
	if err == http.ErrNoCookie {
		http.Error(w, `{"error": "no session"}`, http.StatusUnauthorized)
		return
	}

	sid := cookie.Value
	deleteErr := u.uUsecase.Logout(r.Context(), sid)
	if deleteErr != nil {
		//TODO сделать ошибку в логгер
		return
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
}
