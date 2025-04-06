package http

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"encoding/json"
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
	authMux.HandleFunc("GET /isAuth", h.IsAuth)
	authMux.HandleFunc("POST /logout", h.Logout)
	authMux.HandleFunc("POST /logoutAll", h.LogoutAll)

	r.Handle("/auth/", http.StripPrefix("/auth", authMux))
}

func (h *AuthHandler) IsAuth(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")

	if err != nil || cookie == nil || cookie.Value == "" {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(dto.AuthResponse{UserID: userID, Role: role})
	if err != nil {
		return
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.auth.Logout(cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
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
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := h.auth.LogoutAll(userID, role); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	utils.ClearTokenCookies(w)
	w.WriteHeader(http.StatusOK)
}
