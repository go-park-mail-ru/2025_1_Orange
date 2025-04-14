package utils

import (
	"ResuMatch/internal/usecase"
	"net/http"
	"time"
)

func ClearTokenCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
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
	ctx := r.Context()
	session, err := auth.CreateSession(ctx, userID, role)
	if err != nil {
		WriteAPIError(w, ToAPIError(err))
		return err
	}
	expirationTime := time.Now().Add(time.Duration(86400) * time.Second)
	SetSession(w, session, expirationTime)
	return nil
}

func SetSession(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    value,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		Expires:  expires,
		SameSite: http.SameSiteStrictMode,
	})
}
