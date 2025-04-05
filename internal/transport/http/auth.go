package http

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"net/http"
)

type AuthHandler struct {
	auth usecase.Auth
}

func NewAuthHandler(auth usecase.Auth) AuthHandler {
	return AuthHandler{auth: auth}
}

func (h *AuthHandler) Configure(r *http.ServeMux) {
	authMux := http.NewServeMux()
	authMux.HandleFunc("GET /check-auth", h.CheckAuth)
	authMux.HandleFunc("POST /logout", h.Logout)
	authMux.HandleFunc("POST /logoutAll", h.LogoutAll)

	r.Handle("/auth/", http.StripPrefix("/auth", authMux))
}

func (h *AuthHandler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")

	if err != nil || cookie == nil || cookie.Value == "" {
		utils.NewError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	_, _, err = h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.NewError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.auth.Logout(cookie.Value)
	if err != nil {
		return
	}

	utils.ClearTokenCookies(w)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.NewError(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.auth.LogoutAll(userID, role); err != nil {
		utils.NewError(w, http.StatusInternalServerError, err)
		return
	}

	utils.ClearTokenCookies(w)
	w.WriteHeader(http.StatusOK)
}
