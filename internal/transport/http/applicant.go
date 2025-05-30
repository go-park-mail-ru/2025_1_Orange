package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/metrics"
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"io"
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
	applicantMux.HandleFunc("PUT /profile", h.UpdateProfile)
	applicantMux.HandleFunc("POST /avatar", h.UploadAvatar)
	applicantMux.HandleFunc("POST /emailExists", h.EmailExists)

	r.Handle("/applicant/", http.StripPrefix("/applicant", applicantMux))
}

// Register godoc
// @Tags Applicant
// @Summary Регистрация соискателя
// @Accept json
// @Produce json
// @Param registerData body dto.ApplicantRegister true "Данные для регистрации"
// @Header 200 {string} Set-Cookie "Сессионные cookies"
// @Header 200 {string} X-CSRF-Token "CSRF-токен"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 409 {object} utils.APIError "Пользователь уже существует"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /applicant/register [post]
// @Security csrf_token
func (h *ApplicantHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var registerDTO dto.ApplicantRegister
	if err := utils.ReadJSON(r, &registerDTO); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	applicantID, err := h.applicant.Register(ctx, &registerDTO)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := utils.CreateSession(w, r, h.auth, applicantID, "applicant"); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	middleware.SetCSRFToken(w, r, h.cfg)

	authResp := dto.AuthResponse{UserID: applicantID, Role: "applicant"}
	if err := utils.WriteJSON(w, authResp); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
}

// Login godoc
// @Tags Applicant
// @Summary Авторизация соискателя
// @Description Авторизация соискателя. При успешной авторизации отправляет куки с сессией.
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
// @Router /applicant/login [post]
// @Security csrf_token
func (h *ApplicantHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var loginDTO dto.Login
	if err := utils.ReadJSON(r, &loginDTO); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	applicantID, err := h.applicant.Login(ctx, &loginDTO)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := utils.CreateSession(w, r, h.auth, applicantID, "applicant"); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	middleware.SetCSRFToken(w, r, h.cfg)
	authResp := dto.AuthResponse{UserID: applicantID, Role: "applicant"}
	if err := utils.WriteJSON(w, authResp); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
}

// GetProfile godoc
// @Tags Applicant
// @Summary Получить профиль соискателя
// @Description Возвращает профиль соискателя по ID. Требует авторизации. Доступен только для владельца профиля.
// @Produce json
// @Param id path int true "ID соискателя"
// @Success 200 {object} dto.ApplicantProfileResponse "Профиль соискателя"
// @Failure 400 {object} utils.APIError "Неверный ID"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Нет доступа к этому профилю"
// @Failure 404 {object} utils.APIError "Профиль не найден"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /applicant/profile/{id} [get]
// @Security session_cookie
func (h *ApplicantHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// проверяем сессию
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	_, _, err = h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	requestedID := r.PathValue("id")
	applicantID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	//if applicantID != currentUserID || role != "applicant" {
	//	utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
	//	return
	//}

	applicant, err := h.applicant.GetUser(ctx, applicantID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	//w.Header().Set("Content-Type", "application/json")
	//w.WriteHeader(http.StatusOK)

	if err := utils.WriteJSON(w, applicant); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
}

// UpdateProfile godoc
// @Tags Applicant
// @Summary Обновить профиль соискателя
// @Description Обновляет данные профиля соискателя, кроме аватара. Требует авторизации.
// @Accept json
// @Param updateData body dto.ApplicantProfileUpdate true "Данные для обновления профиля"
// @Success 204
// @Failure 400 {object} utils.APIError "Неверный формат данных"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Нет доступа"
// @Failure 409 {object} utils.APIError "Пользователь уже существует"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /applicant/profile [put]
// @Security csrf_token
// @Security session_cookie
func (h *ApplicantHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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

	if role != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	var applicantDTO dto.ApplicantProfileUpdate
	if err := utils.ReadJSON(r, &applicantDTO); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := h.applicant.UpdateProfile(ctx, userID, &applicantDTO); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadAvatar godoc
// @Tags Applicant
// @Summary Загрузить аватар
// @Description Загружает изображение аватара для профиля соискателя. Требует авторизации и CSRF-токена.
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "Файл изображения (JPEG/PNG, макс. 5MB)"
// @Success 200 {object} dto.UploadStaticResponse "Информация о файле"
// @Failure 400 {object} utils.APIError "Неверный формат файла"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Доступ запрещен"
// @Failure 413 {object} utils.APIError "Файл слишком большой"
// @Failure 500 {object} utils.APIError "Ошибка загрузки файла"
// @Router /applicant/avatar [post]
// @Security session_cookie
// @Security csrf_token
func (h *ApplicantHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// проверяем сессию
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

	if role != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	file, _, err := r.FormFile("avatar")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Applicant Handler", "UploadAvatar").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
	if err = file.Close(); err != nil {
		metrics.LayerErrorCounter.WithLabelValues("Applicant Handler", "UploadAvatar").Inc()
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}

	avatar, err := h.applicant.UpdateAvatar(ctx, userID, data)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := utils.WriteJSON(w, avatar); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
}

// EmailExists godoc
// @Tags Applicant
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
// @Router /applicant/emailExists [post]
// @Security csrf_token
func (h *ApplicantHandler) EmailExists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var emailDTO dto.EmailExistsRequest
	if err := utils.ReadJSON(r, &emailDTO); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	response, err := h.applicant.EmailExists(ctx, emailDTO.Email)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err := utils.WriteJSON(w, response); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
}
