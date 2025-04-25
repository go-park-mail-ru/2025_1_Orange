package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"ResuMatch/pkg/sanitizer"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
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
	resumeMux.HandleFunc("GET /search", h.SearchResumes)

	r.Handle("/resume/", http.StripPrefix("/resume", resumeMux))
}

// CreateResume godoc
// @Tags Resume
// @Summary Создание нового резюме
// @Description Создает новое резюме для авторизованного соискателя. Требует авторизации и CSRF-токена.
// @Accept json
// @Produce json
// @Param resumeData body dto.CreateResumeRequest true "Данные для создания резюме"
// @Success 201 {object} dto.ResumeResponse "Созданное резюме"
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Доступ запрещен (только для соискателей)"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /resume/create [post]
// @Security csrf_token
// @Security session_cookie
func (h *ResumeHandler) CreateResume(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем авторизацию
	cookie, err := r.Cookie("session_id")
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

	// Валидация DTO
	if valid, err := govalidator.ValidateStruct(createResumeRequest); !valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("неверные данные: %v", err))
		return
	}

	// Санитизация всех полей, которые приходят с фронтенда
	createResumeRequest.AboutMe = sanitizer.StrictPolicy.Sanitize(createResumeRequest.AboutMe)
	createResumeRequest.Specialization = sanitizer.StrictPolicy.Sanitize(createResumeRequest.Specialization)
	createResumeRequest.Profession = sanitizer.StrictPolicy.Sanitize(createResumeRequest.Profession) // Дополнение - добавлена санитизация поля профессии
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

// GetResume godoc
// @Tags Resume
// @Summary Получение резюме по ID
// @Description Возвращает полную информацию о резюме по его ID. Доступно всем авторизованным пользователям.
// @Produce json
// @Param id path int true "ID резюме"
// @Success 200 {object} dto.ResumeResponse "Информация о резюме"
// @Failure 400 {object} utils.APIError "Неверный ID"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 404 {object} utils.APIError "Резюме не найдено"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /resume/{id} [get]
// @Security session_cookie
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

// UpdateResume godoc
// @Tags Resume
// @Summary Обновление резюме
// @Description Обновляет информацию о резюме. Доступно только владельцу резюме (соискателю). Требует авторизации и CSRF-токена.
// @Accept json
// @Produce json
// @Param id path int true "ID резюме"
// @Param updateData body dto.UpdateResumeRequest true "Данные для обновления"
// @Success 200 {object} dto.ResumeResponse "Обновленное резюме"
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Доступ запрещен (не владелец)"
// @Failure 404 {object} utils.APIError "Резюме не найдено"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /resume/{id} [put]
// @Security session_cookie
// @Security csrf_token
func (h *ResumeHandler) UpdateResume(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем авторизацию
	cookie, err := r.Cookie("session_id")
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

	// Валидация DTO
	if valid, err := govalidator.ValidateStruct(updateResumeRequest); !valid {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("неверные данные: %v", err))
		return
	}

	updateResumeRequest.AboutMe = sanitizer.StrictPolicy.Sanitize(updateResumeRequest.AboutMe)
	updateResumeRequest.Specialization = sanitizer.StrictPolicy.Sanitize(updateResumeRequest.Specialization)
	updateResumeRequest.Profession = sanitizer.StrictPolicy.Sanitize(updateResumeRequest.Profession) // Дополнение - добавлена санитизация поля профессии
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

// DeleteResume godoc
// @Tags Resume
// @Summary Удаление резюме
// @Description Удаляет резюме по ID. Доступно только владельцу резюме (соискателю). Требует авторизации и CSRF-токена.
// @Produce json
// @Param id path int true "ID резюме"
// @Success 200 {object} dto.DeleteResumeResponse "Результат удаления"
// @Failure 400 {object} utils.APIError "Неверный ID"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Доступ запрещен (не владелец)"
// @Failure 404 {object} utils.APIError "Резюме не найдено"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /resume/{id} [delete]
// @Security session_cookie
// @Security csrf_token
func (h *ResumeHandler) DeleteResume(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем авторизацию
	cookie, err := r.Cookie("session_id")
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

// GetAllResumes godoc
// @Tags Resume
// @Summary Получение всех резюме
// @Description Возвращает список резюме. Для соискателей возвращает только их собственные резюме. Для других ролей - все резюме. Требует авторизации.
// @Produce json
// @Param limit query int false "Количество резюме на странице"
// @Param offset query int false "Смещение от начала списка"
// @Success 200 {object} dto.ResumeShortResponse "Список резюме"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /resume/all [get]
// @Security session_cookie
func (h *ResumeHandler) GetAllResumes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем авторизацию
	cookie, err := r.Cookie("session_id")
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

	// Получаем параметры пагинации из URL
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Значение по умолчанию
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0 // Значение по умолчанию
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	var resumes []dto.ResumeShortResponse

	if role == "applicant" {
		// Получаем список всех резюме соискателя
		resumes, err = h.resume.GetAllResumesByApplicantID(ctx, userID, limit, offset)
		if err != nil {
			utils.WriteAPIError(w, utils.ToAPIError(err))
			return
		}
	} else {
		// Получаем список всех резюме
		resumes, err = h.resume.GetAll(ctx, limit, offset)
		if err != nil {
			utils.WriteAPIError(w, utils.ToAPIError(err))
			return
		}
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resumes); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// SearchResumes godoc
// @Tags Resume
// @Summary Поиск резюме по профессии
// @Description Ищет резюме по профессии. Для соискателей возвращает только их собственные резюме. Для других ролей - все резюме. Требует авторизации.
// @Produce json
// @Param profession query string true "Строка поиска по профессии"
// @Param limit query int false "Количество резюме на странице"
// @Param offset query int false "Смещение от начала списка"
// @Success 200 {array} dto.ResumeShortResponse "Список найденных резюме"
// @Failure 400 {object} utils.APIError "Неверные параметры запроса"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /resume/search [get]
// @Security session_cookie
func (h *ResumeHandler) SearchResumes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Проверяем авторизацию
	cookie, err := r.Cookie("session_id")
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

	// Получаем параметры из URL
	profession := r.URL.Query().Get("profession")
	if profession == "" {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("параметр profession обязателен"))
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Значение по умолчанию
	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	offset := 0 // Значение по умолчанию
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
			return
		}
	}

	// Ищем резюме
	resumes, err := h.resume.SearchResumesByProfession(ctx, userID, role, profession, limit, offset)
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
