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
	"io"
	"net/http"
	"strconv"
)

type EmployerHandler struct {
	auth     usecase.Auth
	employer usecase.Employer
	static   usecase.Static
	cfg      config.CSRFConfig
}

func NewEmployerHandler(auth usecase.Auth, employer usecase.Employer, static usecase.Static, cfg config.CSRFConfig) EmployerHandler {
	return EmployerHandler{auth: auth, employer: employer, static: static, cfg: cfg}
}

func (h *EmployerHandler) Configure(r *http.ServeMux) {
	employerMux := http.NewServeMux()

	employerMux.HandleFunc("POST /register", h.Register)
	employerMux.HandleFunc("POST /login", h.Login)
	employerMux.HandleFunc("GET /profile/{id}", h.GetProfile)
	employerMux.HandleFunc("PUT /profile", h.UpdateProfile)
	employerMux.HandleFunc("POST /logo", h.UploadLogo)
	employerMux.HandleFunc("POST /emailExists", h.EmailExists)

	r.Handle("/employer/", http.StripPrefix("/employer", employerMux))
}

// Register godoc
// @Tags Employer
// @Summary Регистрация работодателя
// @Accept json
// @Produce json
// @Param registerData body dto.EmployerRegister true "Данные для регистрации"
// @Header 200 {string} Set-Cookie "Сессионные cookies"
// @Header 200 {string} X-CSRF-Token "CSRF-токен"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 409 {object} utils.APIError "Пользователь уже существует"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /employer/register [post]
// @Security csrf_token
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

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(dto.AuthResponse{UserID: employerID, Role: "employer"}); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// Login godoc
// @Tags Employer
// @Summary Авторизация работодателя
// @Description Авторизация работодателя. При успешной авторизации отправляет куки с сессией.
// Если пользователь уже авторизован, предыдущие cookies с сессией перезаписываются.
// Также устанавливает CSRF-токен при успешной авторизации.
// @Accept json
// @Produce json
// @Param loginData body dto.Login true "Данные для авторизации (email и пароль)"
// @Header 200 {string} Set-Cookie "Сессионные cookies"
// @Header 200 {string} X-CSRF-Token "CSRF-токен"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 403 {object} utils.APIError "Доступ запрещен (неверные учетные данные)"
// @Failure 404 {object} utils.APIError "Пользователь не найден"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /employer/login [post]
// @Security csrf_token
func (h *EmployerHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var loginDTO dto.Login
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

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(dto.AuthResponse{UserID: employerID, Role: "employer"}); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// GetProfile godoc
// @Tags Employer
// @Summary Получить профиль работодателя
// @Description Возвращает профиль работодателя по ID. Доступен всем.
// @Produce json
// @Param id path int true "ID работодателя"
// @Success 200 {object} dto.EmployerProfileResponse "Профиль работодателя"
// @Failure 400 {object} utils.APIError "Неверный ID"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 404 {object} utils.APIError "Профиль не найден"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /employer/profile/{id} [get]
func (h *EmployerHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestedID := r.PathValue("id")
	employerID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	employer, err := h.employer.GetUser(ctx, employerID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(employer); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateProfile godoc
// @Tags Employer
// @Summary Обновить профиль работодателя
// @Description Обновляет данные профиля работодателя, кроме лого. Требует авторизации.
// @Accept json
// @Param updateData body dto.EmployerProfileUpdate true "Данные для обновления профиля"
// @Success 204
// @Failure 400 {object} utils.APIError "Неверный формат данных"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Нет доступа"
// @Failure 409 {object} utils.APIError "Пользователь уже существует"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /employer/profile [put]
// @Security csrf_token
// @Security session_cookie
func (h *EmployerHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// проверяем сессию
	cookie, err := r.Cookie("session_id")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if role != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	var employerDTO dto.EmployerProfileUpdate
	if err := json.NewDecoder(r.Body).Decode(&employerDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.employer.UpdateProfile(ctx, userID, &employerDTO); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadLogo godoc
// @Tags Employer
// @Summary Загрузить логотип
// @Description Загружает изображение логотипа для профиля работодателя. Требует авторизации и CSRF-токена.
// @Accept multipart/form-data
// @Produce json
// @Param logo formData file true "Файл изображения (JPEG/PNG, макс. 5MB)"
// @Success 200 {object} dto.UploadStaticResponse "Информация о файле"
// @Failure 400 {object} utils.APIError "Неверный формат файла"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Доступ запрещен"
// @Failure 413 {object} utils.APIError "Файл слишком большой"
// @Failure 500 {object} utils.APIError "Ошибка загрузки файла"
// @Router /employer/logo [post]
// @Security session_cookie
// @Security csrf_token
func (h *EmployerHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, role, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if role != "employer" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	file, _, err := r.FormFile("logo")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
	if err = file.Close(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}

	logo, err := h.static.UploadStatic(ctx, data)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err = h.employer.UpdateLogo(ctx, userID, logo.ID); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err = json.NewEncoder(w).Encode(logo); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

// EmailExists godoc
// @Tags Employer
// @Summary Проверка email
// @Description Проверяет, есть ли работодатель с таким email
// @Accept json
// @Produce json
// @Param input body dto.EmailExistsRequest true "Email для проверки"
// @Success 200 {object} dto.EmailExistsResponse
// @Failure 400 {object} utils.APIError
// @Failure 403 {object} utils.APIError
// @Failure 404 {object} utils.APIError
// @Failure 500 {object} utils.APIError
// @Router /employer/emailExists [post]
// @Security csrf_token
func (h *EmployerHandler) EmailExists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var emailDTO dto.EmailExistsRequest
	err := json.NewDecoder(r.Body).Decode(&emailDTO)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	response, err := h.employer.EmailExists(ctx, emailDTO.Email)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(response); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}
