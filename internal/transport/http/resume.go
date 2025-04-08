package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"encoding/json"
	"net/http"
	"strconv"
)

type ResumeHandler struct {
	auth   usecase.Auth
	resume usecase.Resume
	cfg    config.CSRFConfig
}

func NewResumeHandler(auth usecase.Auth, resume usecase.Resume, cfg config.CSRFConfig) ResumeHandler {
	return ResumeHandler{auth: auth, resume: resume, cfg: cfg}
}

func (h *ResumeHandler) Configure(r *http.ServeMux) {
	resumeMux := http.NewServeMux()

	resumeMux.HandleFunc("POST /create", h.CreateResume)
	resumeMux.HandleFunc("GET /{id}", h.GetResume)

	r.Handle("/resume/", http.StripPrefix("/resume", resumeMux))
}

func (h *ResumeHandler) CreateResume(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем авторизацию
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Проверяем, что пользователь - соискатель
	if role != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	// Декодируем запрос
	var createResumeRequest dto.CreateResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&createResumeRequest); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Проверяем, что пользователь создает резюме для себя
	if createResumeRequest.ApplicantID != userID {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	// Создаем резюме
	resume, err := h.resume.Create(ctx, &createResumeRequest)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resume); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *ResumeHandler) GetResume(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Получаем ID резюме из URL
	resumeIDStr := r.PathValue("id")
	resumeID, err := strconv.Atoi(resumeIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Получаем резюме
	resume, err := h.resume.GetByID(ctx, resumeID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resume); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}
