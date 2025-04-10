package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"context"
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
	resumeMux.HandleFunc("PUT /{id}", h.UpdateResume)
	resumeMux.HandleFunc("DELETE /{id}", h.DeleteResume)
	// New handler for employers to view all resumes
	resumeMux.HandleFunc("GET /all", h.GetAllResumes)

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

	// Добавляем ID соискателя в контекст
	ctx = context.WithValue(ctx, "applicantID", userID)

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

func (h *ResumeHandler) UpdateResume(w http.ResponseWriter, r *http.Request) {
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

	// Получаем ID резюме из URL
	resumeIDStr := r.PathValue("id")
	resumeID, err := strconv.Atoi(resumeIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Декодируем запрос
	var updateResumeRequest dto.UpdateResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&updateResumeRequest); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Добавляем ID соискателя в контекст
	ctx = context.WithValue(ctx, "applicantID", userID)

	// Обновляем резюме
	resume, err := h.resume.Update(ctx, resumeID, &updateResumeRequest)
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

func (h *ResumeHandler) DeleteResume(w http.ResponseWriter, r *http.Request) {
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

	// Получаем ID резюме из URL
	resumeIDStr := r.PathValue("id")
	resumeID, err := strconv.Atoi(resumeIDStr)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Удаляем резюме
	response, err := h.resume.Delete(ctx, resumeID, userID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *ResumeHandler) GetAllResumes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем авторизацию
	cookie, err := r.Cookie("session")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	_, role, err := h.auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Проверяем, что пользователь - работодатель
	if role != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	// Получаем список всех резюме
	resumes, err := h.resume.GetAll(ctx)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resumes); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}
