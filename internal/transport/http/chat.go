package http

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"net/http"
	"strconv"
)

type ChatHandler struct {
	auth usecase.Auth
	chat usecase.Chat
}

func NewChatHandler(auth usecase.Auth, chat usecase.Chat) *ChatHandler {
	return &ChatHandler{auth: auth, chat: chat}
}

func (h *ChatHandler) Configure(r *http.ServeMux) {
	chatMux := http.NewServeMux()

	chatMux.HandleFunc("GET /{id}", h.GetChatByID)
	chatMux.HandleFunc("GET /user", h.GetUserChats)
	chatMux.HandleFunc("POST /vacancy/{id}", h.GetVacancyChat)
	chatMux.HandleFunc("GET /{id}/messages", h.GetChatMessages)

	r.Handle("/chat/", http.StripPrefix("/chat", chatMux))
}

// GetVacancyChat godoc
// @Tags Chat
// @Summary Получить чат по вакансии
// @Description Создать или получить существующий чат для вакансии. Только для соискателей (applicant).
// @Param id path int true "ID вакансии"
// @Produce json
// @Success 200 {object} dto.ChatResponse "Чат"
// @Failure 400 {object} utils.APIError "Некорректный запрос"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Доступ запрещён"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /chat/vacancy/{id} [post]
// @Security csrf_token
// @Security session_cookie
func (h *ChatHandler) GetVacancyChat(w http.ResponseWriter, r *http.Request) {
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

	if role != "applicant" {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	requestedID := r.PathValue("id")
	vacancyID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	chat, getErr := h.chat.GetVacancyChat(ctx, vacancyID, userID, role)
	if getErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(getErr))
		return
	}

	if marshalErr := utils.WriteJSON(w, chat); marshalErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(marshalErr))
		return
	}
}

// GetChatByID godoc
// @Tags Chat
// @Summary Получить чат по ID
// @Description Получить чат по его идентификатору. Требует авторизации.
// @Param id path int true "ID чата"
// @Produce json
// @Success 200 {object} dto.ChatResponse "Чат"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 404 {object} utils.APIError "Чат не найден"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /chat/{id} [get]
// @Security csrf_token
// @Security session_cookie
func (h *ChatHandler) GetChatByID(w http.ResponseWriter, r *http.Request) {
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

	requestedID := r.PathValue("id")
	chatID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	chat, getErr := h.chat.GetChat(ctx, chatID, userID, role)
	if getErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(getErr))
		return
	}

	if marshalErr := utils.WriteJSON(w, chat); marshalErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(marshalErr))
		return
	}
}

// GetUserChats godoc
// @Tags Chat
// @Summary Получить чаты пользователя
// @Description Получить все чаты текущего пользователя. Требует авторизации.
// @Produce json
// @Success 200 {object} dto.ChatResponseList "Список чатов"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /chat/user [get]
// @Security csrf_token
// @Security session_cookie
func (h *ChatHandler) GetUserChats(w http.ResponseWriter, r *http.Request) {
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

	chats, getErr := h.chat.GetUserChats(ctx, userID, role)
	if getErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(getErr))
		return
	}

	if marshalErr := utils.WriteJSON(w, chats); marshalErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(marshalErr))
		return
	}
}

// GetChatMessages godoc
// @Tags Chat
// @Summary Получить сообщения чата
// @Description Получить все сообщения из указанного чата. Требует авторизации и доступа к чату.
// @Param id path int true "ID чата"
// @Produce json
// @Success 200 {array} dto.MessageResponse "Список сообщений"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 403 {object} utils.APIError "Доступ запрещён"
// @Failure 404 {object} utils.APIError "Чат не найден"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /chat/{id}/messages [get]
// @Security csrf_token
// @Security session_cookie
func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
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

	requestedID := r.PathValue("id")
	chatID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	chat, getErr := h.chat.GetChat(ctx, chatID, userID, role)
	if getErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(getErr))
		return
	}

	hasAccess := false

	if role == "applicant" {
		if chat.Resume != nil && chat.Resume.ApplicantID == userID {
			hasAccess = true
		}
	} else if role == "employer" {
		if chat.Vacancy != nil && chat.Vacancy.EmployerID == userID {
			hasAccess = true
		}
	}

	if !hasAccess {
		utils.WriteError(w, http.StatusForbidden, entity.ErrForbidden)
		return
	}

	messages, getErr := h.chat.GetChatMessages(ctx, chatID)
	if getErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(getErr))
		return
	}

	if marshalErr := utils.WriteJSON(w, messages); marshalErr != nil {
		utils.WriteAPIError(w, utils.ToAPIError(marshalErr))
		return
	}
}
