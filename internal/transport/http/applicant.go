package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/middleware"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

type ApplicantHandler struct {
	auth      usecase.Auth
	applicant usecase.Applicant
	static    usecase.Static
	cfg       config.CSRFConfig
}

func NewApplicantHandler(auth usecase.Auth, applicant usecase.Applicant, static usecase.Static, cfg config.CSRFConfig) ApplicantHandler {
	return ApplicantHandler{auth: auth, applicant: applicant, static: static, cfg: cfg}
}

func (h *ApplicantHandler) Configure(r *http.ServeMux) {
	applicantMux := http.NewServeMux()

	applicantMux.HandleFunc("POST /register", h.Register)
	applicantMux.HandleFunc("POST /login", h.Login)
	applicantMux.HandleFunc("GET /profile/{id}", h.GetProfile)
	applicantMux.HandleFunc("PUT /profile", h.UpdateProfile)
	applicantMux.HandleFunc("POST /avatar", h.UploadAvatar)

	r.Handle("/applicant/", http.StripPrefix("/applicant", applicantMux))
}

func (h *ApplicantHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var registerDTO dto.ApplicantRegister
	if err := json.NewDecoder(r.Body).Decode(&registerDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
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
}

// Login godoc
// @Tags Авторизация
// @Summary Авторизация пользователя
// @Description Авторизация пользователя. При успешной авторизации отправляет куки с сессией.
// Если пользователь уже авторизован, предыдущие cookies с сессией перезаписываются.
// Также устанавливает CSRF-токен при успешной авторизации.
// @Accept json
// @Produce json
// @Param loginData body dto.ApplicantLogin true "Данные для входа (email и пароль)"
// @Header 200 {string} Set-Cookie "Сессионные cookies"
// @Header 200 {string} X-CSRF-Token "CSRF-токен"
// @Success 200
// @Failure 400 {object} utils.APIError "Неверный формат запроса"
// @Failure 403 {object} utils.APIError "Доступ запрещен (неверные учетные данные)"
// @Failure 404 {object} utils.APIError "Пользователь не найден"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /applicant/login [post]
// @Security _csrf
func (h *ApplicantHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var loginDTO dto.ApplicantLogin
	if err := json.NewDecoder(r.Body).Decode(&loginDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
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
}

func (h *ApplicantHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// проверяем сессию
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	currentUserID, _, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
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

	if applicantID != currentUserID {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	applicant, err := h.applicant.GetUser(ctx, applicantID)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(applicant); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, entity.ErrInternal)
		return
	}
}

func (h *ApplicantHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// проверяем сессию
	cookie, err := r.Cookie("session_id")
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, _, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	var applicantDTO dto.ApplicantProfileUpdate
	if err := json.NewDecoder(r.Body).Decode(&applicantDTO); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.applicant.UpdateProfile(ctx, userID, &applicantDTO); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApplicantHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// проверяем сессию
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie == nil {
		utils.WriteError(w, http.StatusUnauthorized, entity.ErrUnauthorized)
		return
	}

	userID, _, err := h.auth.GetUserIDBySession(ctx, cookie.Value)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	file, _, err := r.FormFile("avatar")
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

	avatarID, err := h.static.UploadStatic(ctx, data)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}

	if err = h.applicant.UpdateAvatar(ctx, userID, avatarID); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
}
