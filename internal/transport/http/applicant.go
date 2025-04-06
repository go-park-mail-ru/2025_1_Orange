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

func (h *ApplicantHandler) Configure(r *http.ServeMux) {
	applicantMux := http.NewServeMux()

	applicantMux.HandleFunc("POST /register", h.Register)
	applicantMux.HandleFunc("POST /login", h.Login)
	applicantMux.HandleFunc("GET /profile/{id}", h.GetProfile)

	r.Handle("/applicant/", http.StripPrefix("/applicant", applicantMux))
}

func (h *ApplicantHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var registerDTO dto.ApplicantRegister
	if err := json.NewDecoder(r.Body).Decode(&registerDTO); err != nil {
		fmt.Println("parsing error")
		utils.NewError(w, http.StatusBadRequest, errors.New("неправильный формат JSON"))
		return
	}

	applicantID, err := h.applicant.Register(ctx, &registerDTO)
	if err != nil {
		var clientErr entity.ClientError
		if errors.As(err, &clientErr) {
			switch {
			case errors.Is(clientErr.Err, entity.ErrBadRequest):
				utils.NewError(w, http.StatusBadRequest, err)
			case errors.Is(clientErr.Err, entity.ErrAlreadyExists):
				utils.NewError(w, http.StatusConflict, err)
			case errors.Is(clientErr.Err, entity.ErrNotFound):
				utils.NewError(w, http.StatusNotFound, err)
			case errors.Is(clientErr.Err, entity.ErrPostgres), errors.Is(clientErr.Err, entity.ErrInternal):
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

func (h *ApplicantHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var loginDTO dto.ApplicantLogin
	if err := json.NewDecoder(r.Body).Decode(&loginDTO); err != nil {
		utils.NewError(w, http.StatusBadRequest, errors.New("неправильный формат JSON"))
		return
	}

	applicantID, err := h.applicant.Login(ctx, &loginDTO)
	if err != nil {
		var clientErr entity.ClientError
		if errors.As(err, &clientErr) {
			switch {
			case errors.Is(clientErr.Err, entity.ErrForbidden):
				utils.NewError(w, http.StatusForbidden, err)
			case errors.Is(clientErr.Err, entity.ErrNotFound):
				utils.NewError(w, http.StatusNotFound, err)
			case errors.Is(clientErr.Err, entity.ErrPostgres), errors.Is(clientErr.Err, entity.ErrInternal):
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

func (h *ApplicantHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// проверяем сессию
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.NewError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, _, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.NewError(w, http.StatusUnauthorized, err)
		return
	}

	//requestedID := r.URL.Query().Get("id")
	requestedID := r.PathValue("id")
	applicantID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.NewError(w, http.StatusBadRequest, errors.New("неверный id"))
		return
	}

	if applicantID != currentUserID {
		utils.NewError(w, http.StatusForbidden, errors.New("нет доступа"))
		return
	}

	applicant, err := h.applicant.GetUser(ctx, applicantID)
	if err != nil {
		var clientErr entity.ClientError
		if errors.As(err, &clientErr) {
			switch {
			case errors.Is(clientErr.Err, entity.ErrNotFound):
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
