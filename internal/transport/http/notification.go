package http

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase"
	"net/http"
	"strconv"
)

type NotificationHandler struct {
	notification usecase.Notification
	auth         usecase.Auth
}

func NewNotificationHandler(
	notification usecase.Notification,
	auth usecase.Auth,
) NotificationHandler {
	return NotificationHandler{
		notification: notification,
		auth:         auth,
	}
}

func (h *NotificationHandler) Configure(r *http.ServeMux) {
	notificationMux := http.NewServeMux()

	notificationMux.HandleFunc("GET /user", h.GetNotificationsForUser)
	notificationMux.HandleFunc("PUT /read/{id}", h.ReadNotification)
	notificationMux.HandleFunc("PUT /readAll", h.ReadAllNotifications)
	notificationMux.HandleFunc("DELETE /clear", h.DeleteAllNotifications)
	r.Handle("/notification/", http.StripPrefix("/notification", notificationMux))
}

// GetNotificationsForUser godoc
// @Tags Notification
// @Summary Получить все уведомления пользователя
// @Description Получаем список уведомлений для соискателя или работодателя. Требует авторизации.
// @Produce json
// @Success 200 {object} entity.NotificationsList "Список уведомлений"
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /notification/user [get]
// @Security session_cookie
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

// ReadNotification godoc
// @Tags Notification
// @Summary Прочитать уведомление
// @Description Прочитать конкретное уведомление по id. Требует авторизации.
// @Produce json
// @Param id path int true "ID уведомления"
// @Success 200
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /notification/read/{id} [put]
// @Security csrf_token
// @Security session_cookie
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

// ReadAllNotifications godoc
// @Tags Notification
// @Summary Прочитать все уведомления
// @Description Прочитать все уведомления пользователя. Требует авторизации.
// @Produce json
// @Success 200
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /notification/readAll [put]
// @Security csrf_token
// @Security session_cookie
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

// DeleteAllNotifications godoc
// @Tags Notification
// @Summary Удалить все уведомления
// @Description Удалить все уведомления пользователя. Требует авторизации.
// @Produce json
// @Success 200
// @Failure 401 {object} utils.APIError "Не авторизован"
// @Failure 500 {object} utils.APIError "Внутренняя ошибка сервера"
// @Router /notification/clear [delete]
// @Security csrf_token
// @Security session_cookie
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
