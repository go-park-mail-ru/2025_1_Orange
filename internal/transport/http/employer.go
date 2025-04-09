package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type EmployerHandler struct {
	auth     usecase.Auth
	employer usecase.Employer
	cfg      config.CSRFConfig
}

func NewEmployerHandler(auth usecase.Auth, employer usecase.Employer, cfg config.CSRFConfig) EmployerHandler {
	return EmployerHandler{auth: auth, employer: employer, cfg: cfg}
}

func (h *EmployerHandler) Configure(r *http.ServeMux) {
	employerMux := http.NewServeMux()

	employerMux.HandleFunc("POST /register", h.Register)
	employerMux.HandleFunc("POST /login", h.Login)
	employerMux.HandleFunc("GET /profile/{id}", h.GetProfile)

	r.Handle("/employer/", http.StripPrefix("/employer", employerMux))
}

func (h *EmployerHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var registerDTO dto.EmployerRegister
	if err := json.NewDecoder(r.Body).Decode(&registerDTO); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrBadRequest)
		return
	}
	fmt.Printf("comp=%s, addr=%s", registerDTO.CompanyName, registerDTO.LegalAddress)
	employerID, err := h.employer.Register(ctx, &registerDTO)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := utils.CreateSession(w, r, h.auth, employerID, "employer"); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
	middleware.SetCSRFToken(w, r, h.cfg)
}

func (h *EmployerHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var loginDTO dto.EmployerLogin
	if err := json.NewDecoder(r.Body).Decode(&loginDTO); err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrBadRequest)
		return
	}

	employerID, err := h.employer.Login(ctx, &loginDTO)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := utils.CreateSession(w, r, h.auth, employerID, "employer"); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
	middleware.SetCSRFToken(w, r, h.cfg)
}

func (h *EmployerHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// проверяем сессию
	cookie, err := r.Cookie("session")
	if err != nil || cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, _, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	//requestedID := r.URL.Query().Get("id")
	requestedID := r.PathValue("id")
	employerID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	if employerID != currentUserID {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	employer, err := h.employer.GetUser(ctx, employerID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(employer); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}
