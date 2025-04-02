package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
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

func (h *EmployerHandler) Configure(r *httprouter.Router) {
	r.POST("/api/v1/employer/register", h.Register)
	r.POST("/api/v1/employer/login", h.Login)
	r.GET("/api/v1/employer/profile", h.GetProfile)
}

func (h *EmployerHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var registerDTO dto.EmployerRegister
	if err := json.NewDecoder(r.Body).Decode(&registerDTO); err != nil {
		utils.NewError(w, http.StatusBadRequest, errors.New("неправильный формат JSON"))
		return
	}
	fmt.Printf("comp=%s, addr=%s", registerDTO.CompanyName, registerDTO.LegalAddress)
	employerID, err := h.employer.Register(&registerDTO)
	if err != nil {
		var clientErr entity.ClientError
		if errors.As(err, &clientErr) {
			switch {
			case errors.Is(clientErr.Data, entity.ErrBadRequest):
				utils.NewError(w, http.StatusBadRequest, err)
			case errors.Is(clientErr.Data, entity.ErrAlreadyExists):
				utils.NewError(w, http.StatusConflict, err)
			case errors.Is(clientErr.Data, entity.ErrNotFound):
				utils.NewError(w, http.StatusNotFound, err)
			case errors.Is(clientErr.Data, entity.ErrPostgres), errors.Is(clientErr.Data, entity.ErrInternal):
				utils.NewError(w, http.StatusInternalServerError, err)
			default:
				utils.NewError(w, http.StatusInternalServerError, err)
			}

		} else {
			utils.NewError(w, http.StatusInternalServerError, err)
		}
		return
	}

	if err := utils.CreateSession(w, r, h.auth, employerID, "employer"); err != nil {
		utils.NewError(w, http.StatusInternalServerError, err)
		return
	}
	middleware.SetCSRFToken(w, r, h.cfg)
}

func (h *EmployerHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var loginDTO dto.EmployerLogin
	if err := json.NewDecoder(r.Body).Decode(&loginDTO); err != nil {
		utils.NewError(w, http.StatusBadRequest, errors.New("неправильный формат JSON"))
		return
	}

	employerID, err := h.employer.Login(&loginDTO)
	if err != nil {
		var clientErr entity.ClientError
		if errors.As(err, &clientErr) {
			switch {
			case errors.Is(clientErr.Data, entity.ErrForbidden):
				utils.NewError(w, http.StatusForbidden, err)
			case errors.Is(clientErr.Data, entity.ErrNotFound):
				utils.NewError(w, http.StatusNotFound, err)
			case errors.Is(clientErr.Data, entity.ErrPostgres), errors.Is(clientErr.Data, entity.ErrInternal):
				utils.NewError(w, http.StatusInternalServerError, err)
			default:
				utils.NewError(w, http.StatusInternalServerError, err)
			}

		} else {
			utils.NewError(w, http.StatusInternalServerError, err)
		}
		return
	}

	if err := utils.CreateSession(w, r, h.auth, employerID, "employer"); err != nil {
		utils.NewError(w, http.StatusInternalServerError, err)
		return
	}
	middleware.SetCSRFToken(w, r, h.cfg)
}

func (h *EmployerHandler) GetProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// проверяем сессию
	currentUserID, _, err := utils.GetUserIDFromSession(r, h.auth)
	if err != nil {
		utils.NewError(w, http.StatusUnauthorized, err)
		return
	}

	requestedID := r.URL.Query().Get("id")
	employerID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.NewError(w, http.StatusBadRequest, errors.New("неверный id"))
		return
	}

	if employerID != currentUserID {
		utils.NewError(w, http.StatusForbidden, errors.New("нет доступа"))
		return
	}

	employer, err := h.employer.GetUser(employerID)
	if err != nil {
		var clientErr entity.ClientError
		if errors.As(err, &clientErr) {
			switch {
			case errors.Is(clientErr.Data, entity.ErrNotFound):
				utils.NewError(w, http.StatusNotFound, err)
			default:
				utils.NewError(w, http.StatusInternalServerError, err)
			}
		} else {
			utils.NewError(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(employer)
	if err != nil {
		return
	}
}
