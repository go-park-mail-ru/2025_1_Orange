package utils

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/usecase"
	"net/http"
	"time"
)

func GetUserIDFromSession(r *http.Request, auth usecase.Auth) (int, string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return 0, "", entity.NewClientError("отсутствует cookies с сессией", entity.ErrUnauthorized)
	}

	userID, role, err := auth.GetUserIDBySession(cookie.Value)
	if err != nil {
		return 0, "", entity.NewClientError("необходима авторизация", entity.ErrUnauthorized)
	}

	return userID, role, nil
}

func NullifyTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(-24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(-24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

func CreateSession(w http.ResponseWriter, r *http.Request, auth usecase.Auth, userID int, role string) error {
	session, err := auth.CreateSession(userID, role)
	if err != nil {
		NewError(w, http.StatusInternalServerError, err)
		return err
	}
	// TODO добавить SessionAliveTime в конфиг
	expirationTime := time.Now().Add(time.Duration(86400) * time.Second)
	SetSession(w, session, expirationTime)
	return nil
}

func SetSession(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    value,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		Expires:  expires,
		SameSite: http.SameSiteStrictMode,
	})
}
