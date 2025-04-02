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

type ApplicantHandler struct {
	auth      usecase.Auth
	applicant usecase.Applicant
	cfg       config.CSRFConfig
}

func NewApplicantHandler(auth usecase.Auth, applicant usecase.Applicant, cfg config.CSRFConfig) ApplicantHandler {
	return ApplicantHandler{auth: auth, applicant: applicant, cfg: cfg}
}

func (h *ApplicantHandler) Configure(r *httprouter.Router) {
	r.POST("/api/v1/applicant/register", h.Register)
	r.POST("/api/v1/applicant/login", h.Login)
	r.GET("/api/v1/applicant/profile", h.GetProfile)
}

func (h *ApplicantHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var registerDTO dto.ApplicantRegister
	if err := json.NewDecoder(r.Body).Decode(&registerDTO); err != nil {
		fmt.Println("parsing error")
		utils.NewError(w, http.StatusBadRequest, errors.New("неправильный формат JSON"))
		return
	}

	applicantID, err := h.applicant.Register(&registerDTO)
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

	if err := utils.CreateSession(w, r, h.auth, applicantID, "applicant"); err != nil {
		utils.NewError(w, http.StatusInternalServerError, err)
		return
	}

	middleware.SetCSRFToken(w, r, h.cfg)
}

func (h *ApplicantHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var loginDTO dto.ApplicantLogin
	if err := json.NewDecoder(r.Body).Decode(&loginDTO); err != nil {
		utils.NewError(w, http.StatusBadRequest, errors.New("неправильный формат JSON"))
		return
	}

	applicantID, err := h.applicant.Login(&loginDTO)
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

	if err := utils.CreateSession(w, r, h.auth, applicantID, "applicant"); err != nil {
		utils.NewError(w, http.StatusInternalServerError, err)
		return
	}
	middleware.SetCSRFToken(w, r, h.cfg)
}

func (h *ApplicantHandler) GetProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// проверяем сессию
	currentUserID, _, err := utils.GetUserIDFromSession(r, h.auth)
	if err != nil {
		utils.NewError(w, http.StatusUnauthorized, err)
		return
	}

	requestedID := r.URL.Query().Get("id")
	applicantID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.NewError(w, http.StatusBadRequest, errors.New("неверный id"))
		return
	}

	if applicantID != currentUserID {
		utils.NewError(w, http.StatusForbidden, errors.New("нет доступа"))
		return
	}

	applicant, err := h.applicant.GetUser(applicantID)
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
	err = json.NewEncoder(w).Encode(applicant)
	if err != nil {
		return
	}
}
