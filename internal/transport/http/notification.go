package http

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/transport/ws"
	"ResuMatch/internal/usecase"
	"net/http"
	"strconv"
)

type NotificationHandler struct {
	notification usecase.Notification
	auth         usecase.Auth
	wsPool       *ws.WebsocketPool
}

func NewNotificationHandler(
	notification usecase.Notification,
	auth usecase.Auth,
	wsPool *ws.WebsocketPool,

) NotificationHandler {
	return NotificationHandler{
		notification: notification,
		auth:         auth,
		wsPool:       wsPool,
	}
}

func (h *NotificationHandler) Configure(r *http.ServeMux) {
	notificationMux := http.NewServeMux()

	notificationMux.HandleFunc("GET /user", h.GetNotificationsForUser)
	notificationMux.HandleFunc("PUT /read/{id}", h.ReadNotification)
	notificationMux.HandleFunc("PUT /readAll", h.ReadAllNotifications)
	notificationMux.HandleFunc("DELETE /clear", h.DeleteAllNotifications)
	notificationMux.HandleFunc("GET /ws", h.wsPool.Connect)
	//r.Handle("/notification/ws", http.HandlerFunc(h.wsPool.Connect))
	r.Handle("/notification/", http.StripPrefix("/notification", notificationMux))
}

func (h *NotificationHandler) GetNotificationsForUser(w http.ResponseWriter, r *http.Request) {
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

	notifications, err := h.notification.GetNotificationsForUser(ctx, userID, role)
	if err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
	resp := entity.NotificationsList(notifications)
	if err := utils.WriteJSON(w, resp); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *NotificationHandler) ReadNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	notificationID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, entity.ErrBadRequest)
		return
	}

	if err := h.notification.ReadNotification(ctx, notificationID, userID); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *NotificationHandler) ReadAllNotifications(w http.ResponseWriter, r *http.Request) {
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

	if err := h.notification.ReadAllNotifications(ctx, userID, role); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *NotificationHandler) DeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
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

	if err := h.notification.DeleteAllNotifications(ctx, userID, role); err != nil {
		utils.WriteAPIError(w, utils.ToAPIError(err))
		return
	}
	w.WriteHeader(http.StatusOK)
}
