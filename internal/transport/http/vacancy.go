package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"ResuMatch/pkg/sanitizer"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type VacancyHandler struct {
	auth    usecase.Auth
	vacancy usecase.Vacancy
	cfg     config.CSRFConfig
}

func NewVacancyHandler(auth usecase.Auth, vac usecase.Vacancy, cfg config.CSRFConfig) VacancyHandler {
	return VacancyHandler{
		auth:    auth,
		vacancy: vac,
		cfg:     cfg,
	}
}

func (h *VacancyHandler) Configure(r *http.ServeMux) {
	vacancyMux := http.NewServeMux()
	vacancyMux.HandleFunc("GET /vacancies", h.GetAllVacancies)
	vacancyMux.HandleFunc("POST /vacancies", h.CreateVacancy)
	vacancyMux.HandleFunc("GET /vacancy/{id}", h.GetVacancy)
	vacancyMux.HandleFunc("PUT /vacancy/{id}", h.UpdateVacancy)
	vacancyMux.HandleFunc("DELETE /vacancy/{id}", h.DeleteVacancy)
	vacancyMux.HandleFunc("POST /vacancy/{id}/response", h.ApplyToVacancy)
	vacancyMux.HandleFunc("GET /employer/{id}/vacancies", h.GetActiveVacanciesByEmployer)
	vacancyMux.HandleFunc("GET /applicant/{id}/vacancies", h.GetVacanciesByApplicant)
	vacancyMux.HandleFunc("GET /search", h.SearchVacancies)
	vacancyMux.HandleFunc("POST /search/specializations", h.SearchVacanciesBySpecializations)
	vacancyMux.HandleFunc("POST /search/combined", h.SearchVacanciesByQueryAndSpecializations)
	vacancyMux.HandleFunc("GET /applicant/{id}/liked", h.GetLikedVacancies)
	vacancyMux.HandleFunc("POST /vacancy/{id}/like", h.LikeVacancy)
	vacancyMux.HandleFunc("GET /vacancy/{id}/response/list", h.GetResponsesOnVacancy)
	r.Handle("/vacancy/", http.StripPrefix("/vacancy", vacancyMux))
}

func (h *VacancyHandler) CreateVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "CreateVacancy").Inc()
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	var vacancyCreate dto.VacancyCreate
	if err := json.NewDecoder(r.Body).Decode(&vacancyCreate); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "CreateVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	vacancyCreate.Title = sanitizer.StrictPolicy.Sanitize(vacancyCreate.Title)
	vacancyCreate.Specialization = sanitizer.StrictPolicy.Sanitize(vacancyCreate.Specialization)
	vacancyCreate.WorkFormat = sanitizer.StrictPolicy.Sanitize(vacancyCreate.WorkFormat)
	vacancyCreate.Employment = sanitizer.StrictPolicy.Sanitize(vacancyCreate.Employment)
	vacancyCreate.Schedule = sanitizer.StrictPolicy.Sanitize(vacancyCreate.Schedule)
	vacancyCreate.City = sanitizer.StrictPolicy.Sanitize(vacancyCreate.City)
	vacancyCreate.Description = sanitizer.StrictPolicy.Sanitize(vacancyCreate.Description)
	vacancyCreate.Experience = sanitizer.StrictPolicy.Sanitize(vacancyCreate.Experience)
	vacancyCreate.Tasks = sanitizer.StrictPolicy.Sanitize(vacancyCreate.Tasks)
	vacancyCreate.Requirements = sanitizer.StrictPolicy.Sanitize(vacancyCreate.Requirements)
	vacancyCreate.OptionalRequirements = sanitizer.StrictPolicy.Sanitize(vacancyCreate.OptionalRequirements)

	for i, skill := range vacancyCreate.Skills {
		vacancyCreate.Skills[i] = sanitizer.StrictPolicy.Sanitize(skill)
	}

	vacancy, err := h.vacancy.CreateVacancy(ctx, currentUserID, &vacancyCreate)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "CreateVacancy").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) GetVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID = 0
	var userRole string

	cookie, err := r.Cookie("session_id")
	if err == nil && cookie != nil {
		currentUserID, currenyUserRole, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
		if err == nil {
			userID = currentUserID
			userRole = currenyUserRole
		}
	}

	idStr := r.PathValue("id")
	vacancyID, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	vacancy, err := h.vacancy.GetVacancy(ctx, vacancyID, userID, userRole)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacancy").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session_id")
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "UpdateVacancy").Inc()
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	idStr := r.PathValue("id")
	vacancyID, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "UpdateVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	var vacancyUpdate dto.VacancyUpdate
	if err := json.NewDecoder(r.Body).Decode(&vacancyUpdate); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "UpdateVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Санитизация всех строковых полей
	vacancyUpdate.Title = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.Title)
	vacancyUpdate.Description = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.Description)
	vacancyUpdate.WorkFormat = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.WorkFormat)
	vacancyUpdate.Employment = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.Employment)
	vacancyUpdate.Schedule = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.Schedule)
	vacancyUpdate.City = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.City)
	vacancyUpdate.Experience = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.Experience)
	vacancyUpdate.Specialization = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.Specialization)
	vacancyUpdate.Requirements = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.Requirements)
	vacancyUpdate.OptionalRequirements = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.OptionalRequirements)
	vacancyUpdate.Tasks = sanitizer.StrictPolicy.Sanitize(vacancyUpdate.Tasks)

	// Санитизация массивов строк

	for i, skill := range vacancyUpdate.Skills {
		vacancyUpdate.Skills[i] = sanitizer.StrictPolicy.Sanitize(skill)
	}

	vacancy, err := h.vacancy.UpdateVacancy(ctx, vacancyID, currentUserID, &vacancyUpdate)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancy); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "UpdateVacancy").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) DeleteVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверка сессии
	cookie, err := r.Cookie("session_id")
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "DeleteVacancy").Inc()
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	if cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "DeleteVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	response, err := h.vacancy.DeleteVacancy(ctx, id, currentUserID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "DeleteVacancy").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) GetAllVacancies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID int
	var userRole string

	cookie, err := r.Cookie("session_id")
	if err == nil && cookie != nil {
		currentUserID, currentUserRole, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
		if err == nil {
			userID = currentUserID
			userRole = currentUserRole
		}
	}

	// Получаем параметры пагинации из URL
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Значение по умолчанию
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetAllVacancies").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0 // Значение по умолчанию
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetAllVacancies").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	vacancies, err := h.vacancy.GetAll(ctx, userID, userRole, limit, offset)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetAllVacancies").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) ApplyToVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "ApplyToVacancy").Inc()
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	// Получаем ID вакансии из URL
	vacancyID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "ApplyToVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Получаем ID текущего пользователя
	applicantID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil || userType != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	err = h.vacancy.ApplyToVacancy(ctx, vacancyID, applicantID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *VacancyHandler) GetResponsesOnVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("id")
	vacancyID, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	_, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetActiveVacanciesByEmployer").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetActiveVacanciesByEmployer").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	resumes, err := h.vacancy.GetRespondedResumeOnVacancy(ctx, vacancyID, limit, offset)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resumes); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Resume Handler", "GetAllResumes").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) GetActiveVacanciesByEmployer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID = 0
	var userRole string

	idStr := r.PathValue("id")
	employerID, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetActiveVacanciesByEmployer").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err == nil && cookie != nil {
		currentUserID, currenyUserRole, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
		if err == nil {
			userID = currentUserID
			userRole = currenyUserRole
		}
	}

	// Получаем параметры пагинации из URL
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Значение по умолчанию
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetActiveVacanciesByEmployer").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0 // Значение по умолчанию
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetActiveVacanciesByEmployer").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	vacancies, err := h.vacancy.GetActiveVacanciesByEmployerID(ctx, employerID, userID, userRole, limit, offset)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetActiveVacanciesByEmployer").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) GetVacanciesByApplicant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	idStr := r.PathValue("id")
	applicantID, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	if applicantID != currentUserID {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	// Получаем параметры пагинации из URL
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Значение по умолчанию
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0 // Значение по умолчанию
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	// Получаем список вакансий, на которые откликнулся соискатель
	vacancies, err := h.vacancy.GetVacanciesByApplicantID(ctx, applicantID, limit, offset)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// SearchVacancies godoc
// @Tags Vacancy
// @Summary Поиск вакансий
// @Description Ищет вакансии по заданному запросу. Поиск выполняется по названию должности, специализации и названию компании. Для работодателей возвращает только их собственные вакансии. Для других ролей - все вакансии.
// @Produce json
// @Param query query string true "Строка поиска"
// @Param limit query int false "Количество вакансий на странице"
// @Param offset query int false "Смещение от начала списка"
// @Success 200 {array} dto.VacancyShortResponse "Список найденных вакансий"
// @Failure 400 {object} utils.APIError "Неверные параметры запроса"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /vacancy/search [get]
func (h *VacancyHandler) SearchVacancies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID = 0
	var userRole string

	// Проверяем авторизацию (если есть)
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie != nil {
		currentUserID, currentUserRole, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
		if err == nil {
			userID = currentUserID
			userRole = currentUserRole
		}
	}

	// Получаем параметры из URL
	searchQuery := r.URL.Query().Get("query")
	if searchQuery == "" {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacancies").Inc()
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("параметр query обязателен"))
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Значение по умолчанию
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacancies").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0 // Значение по умолчанию
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacancies").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	// Ищем вакансии
	vacancies, err := h.vacancy.SearchVacancies(ctx, userID, userRole, searchQuery, limit, offset)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacancies").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// SearchVacanciesBySpecializations обрабатывает запрос на поиск вакансий по специализациям
func (h *VacancyHandler) SearchVacanciesBySpecializations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID = 0
	var userRole string

	// Проверяем авторизацию (если есть)
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie != nil {
		currentUserID, currentUserRole, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
		if err == nil {
			userID = currentUserID
			userRole = currentUserRole
		}
	}

	// Получаем параметры пагинации из URL
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Значение по умолчанию
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesBySpecializations").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0 // Значение по умолчанию
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesBySpecializations").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	// Декодируем тело запроса
	var searchRequest dto.SearchBySpecializationsRequest
	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesBySpecializations").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Санитизация входных данных
	for i, spec := range searchRequest.Specializations {
		searchRequest.Specializations[i] = sanitizer.StrictPolicy.Sanitize(spec)
	}

	// Если список специализаций пуст, возвращаем ошибку
	if len(searchRequest.Specializations) == 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("список специализаций не может быть пустым"))
		return
	}

	// Ищем вакансии по специализациям
	vacancies, err := h.vacancy.SearchVacanciesBySpecializations(ctx, userID, userRole, searchRequest.Specializations, limit, offset)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesBySpecializations").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// SearchVacanciesByQueryAndSpecializations обрабатывает запрос на комбинированный поиск вакансий
func (h *VacancyHandler) SearchVacanciesByQueryAndSpecializations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID = 0
	var userRole string

	// Проверяем авторизацию (если есть)
	cookie, err := r.Cookie("session_id")
	if err == nil && cookie != nil {
		currentUserID, currentUserRole, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
		if err == nil {
			userID = currentUserID
			userRole = currentUserRole
		}
	}

	// Получаем параметр поиска из URL
	searchQuery := r.URL.Query().Get("query")
	if searchQuery == "" {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesByQueryAndSpecializations").Inc()
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("параметр query обязателен"))
		return
	}

	// Получаем параметры пагинации из URL
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Значение по умолчанию
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesByQueryAndSpecializations").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0 // Значение по умолчанию
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesByQueryAndSpecializations").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	// Декодируем тело запроса для получения списка специализаций
	var searchRequest dto.SearchByQueryAndSpecializationsRequest
	if err := json.NewDecoder(r.Body).Decode(&searchRequest); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesByQueryAndSpecializations").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	// Санитизация входных данных
	for i, spec := range searchRequest.Specializations {
		searchRequest.Specializations[i] = sanitizer.StrictPolicy.Sanitize(spec)
	}

	// Если список специализаций пуст, возвращаем ошибку
	if len(searchRequest.Specializations) == 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("список специализаций не может быть пустым"))
		return
	}

	// Выполняем комбинированный поиск вакансий
	vacancies, err := h.vacancy.SearchVacanciesByQueryAndSpecializations(ctx, userID, userRole, searchQuery, searchRequest.Specializations, limit, offset)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "SearchVacanciesByQueryAndSpecializations").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) GetLikedVacancies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, userType, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if userType != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	idStr := r.PathValue("id")
	applicantID, err := strconv.Atoi(idStr)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	if applicantID != currentUserID {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetVacanciesByApplicant").Inc()
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	if applicantID != currentUserID {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	vacancies, err := h.vacancy.GetLikedVacancies(ctx, applicantID, limit, offset)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(vacancies); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "GetLikedVacancies").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *VacancyHandler) LikeVacancy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "LikeVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	currentUserID, currentUserRole, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil || currentUserRole != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	vacancyID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Vacancy Handler", "LikeVacancy").Inc()
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	err = h.vacancy.LikeVacancy(ctx, vacancyID, currentUserID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
