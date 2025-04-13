package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"ResuMatch/pkg/sanitizer"
	"encoding/json"
	"net/http"
	"strconv"
)

type ResumeHandler struct {
	auth   usecase.Auth
	resume usecase.ResumeUsecase
	cfg    config.CSRFConfig
}

func NewResumeHandler(auth usecase.Auth, resume usecase.ResumeUsecase, cfg config.CSRFConfig) ResumeHandler {
	return ResumeHandler{auth: auth, resume: resume, cfg: cfg}
}

func (h *ResumeHandler) Configure(r *http.ServeMux) {
	resumeMux := http.NewServeMux()

	resumeMux.HandleFunc("POST /create", h.CreateResume)
	resumeMux.HandleFunc("GET /{id}", h.GetResume)
	resumeMux.HandleFunc("PUT /{id}", h.UpdateResume)
	resumeMux.HandleFunc("DELETE /{id}", h.DeleteResume)
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

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
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

	// Санитизация всех полей, которые приходят с фронтенда
	createResumeRequest.AboutMe = sanitizer.StrictPolicy.Sanitize(createResumeRequest.AboutMe)
	createResumeRequest.Specialization = sanitizer.StrictPolicy.Sanitize(createResumeRequest.Specialization)
	createResumeRequest.EducationalInstitution = sanitizer.StrictPolicy.Sanitize(createResumeRequest.EducationalInstitution)

	// Санитизация массивов строк
	for i, skill := range createResumeRequest.Skills {
		createResumeRequest.Skills[i] = sanitizer.StrictPolicy.Sanitize(skill)
	}

	for i, spec := range createResumeRequest.AdditionalSpecializations {
		createResumeRequest.AdditionalSpecializations[i] = sanitizer.StrictPolicy.Sanitize(spec)
	}

	// Санитизация опыта работы
	for i, we := range createResumeRequest.WorkExperiences {
		createResumeRequest.WorkExperiences[i].EmployerName = sanitizer.StrictPolicy.Sanitize(we.EmployerName)
		createResumeRequest.WorkExperiences[i].Position = sanitizer.StrictPolicy.Sanitize(we.Position)
		createResumeRequest.WorkExperiences[i].Duties = sanitizer.StrictPolicy.Sanitize(we.Duties)
		createResumeRequest.WorkExperiences[i].Achievements = sanitizer.StrictPolicy.Sanitize(we.Achievements)
	}

	resume, err := h.resume.Create(ctx, userID, &createResumeRequest)
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

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
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

	updateResumeRequest.AboutMe = sanitizer.StrictPolicy.Sanitize(updateResumeRequest.AboutMe)
	updateResumeRequest.Specialization = sanitizer.StrictPolicy.Sanitize(updateResumeRequest.Specialization)
	updateResumeRequest.EducationalInstitution = sanitizer.StrictPolicy.Sanitize(updateResumeRequest.EducationalInstitution)

	// Санитизация массивов строк
	for i, skill := range updateResumeRequest.Skills {
		updateResumeRequest.Skills[i] = sanitizer.StrictPolicy.Sanitize(skill)
	}

	for i, spec := range updateResumeRequest.AdditionalSpecializations {
		updateResumeRequest.AdditionalSpecializations[i] = sanitizer.StrictPolicy.Sanitize(spec)
	}

	// Санитизация опыта работы
	for i, we := range updateResumeRequest.WorkExperiences {
		updateResumeRequest.WorkExperiences[i].EmployerName = sanitizer.StrictPolicy.Sanitize(we.EmployerName)
		updateResumeRequest.WorkExperiences[i].Position = sanitizer.StrictPolicy.Sanitize(we.Position)
		updateResumeRequest.WorkExperiences[i].Duties = sanitizer.StrictPolicy.Sanitize(we.Duties)
		updateResumeRequest.WorkExperiences[i].Achievements = sanitizer.StrictPolicy.Sanitize(we.Achievements)
	}

	resume, err := h.resume.Update(ctx, resumeID, userID, &updateResumeRequest)
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

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
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

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	_, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
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
