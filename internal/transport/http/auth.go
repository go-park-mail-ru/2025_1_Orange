package http

import (
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

type AuthHandler struct {
	auth usecase.Auth
}

func NewAuthHandler(auth usecase.Auth) AuthHandler {
	return AuthHandler{auth: auth}
}

func (h *AuthHandler) Configure(r *httprouter.Router) {
	r.GET("/api/v1/auth/check-auth", h.CheckAuth)
	r.POST("/api/v1/auth/logout", h.Logout)
	r.POST("/api/v1/auth/logoutAll", h.LogoutAll)
}

func (h *AuthHandler) CheckAuth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if _, _, err := utils.GetUserIDFromSession(r, h.auth); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
