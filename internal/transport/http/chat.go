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

func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	chatID, err := strconv.Atoi(requestedID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
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
