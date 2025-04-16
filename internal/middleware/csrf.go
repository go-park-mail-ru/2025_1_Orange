package middleware

import (
	"ResuMatch/internal/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"time"
)

func generateToken(r *http.Request, sessionID string, cfg config.CSRFConfig) string {
	h := hmac.New(sha256.New, []byte(cfg.Secret))
	if sessionID != "" {
		h.Write([]byte(sessionID))
	} else {
		// Соль в виде ip адреса
		h.Write([]byte(r.RemoteAddr))
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func SetCSRFToken(w http.ResponseWriter, r *http.Request, cfg config.CSRFConfig) {
	sessionCookie, _ := r.Cookie("session_id")
	var sessionID string
	if sessionCookie != nil {
		sessionID = sessionCookie.Value
	}

	token := generateToken(r, sessionID, cfg)

	http.SetCookie(w, &http.Cookie{
		Name:  cfg.CookieName,
		Value: token,
		// устанавливаем куку для всех маршрутов домена
		Path:     "/",
		Secure:   cfg.Secure,
		HttpOnly: cfg.HttpOnly,
		Expires:  time.Now().Add(cfg.Lifetime),
		SameSite: parsedSameSite(cfg),
	})
	w.Header().Set("X-CSRF-Token", token)
}

func CSRFMiddleware(cfg config.CSRFConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, является ли это запрос статики
			isStatic := strings.HasPrefix(r.URL.Path, "/api/v1/static/")

			// Исключаем безопасные методы
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				if !isStatic {
					SetCSRFToken(w, r, cfg)
				}

				next.ServeHTTP(w, r)
				return
			}

			// Проверяем токен для небезопасных методов
			receivedToken := r.Header.Get("X-CSRF-Token")
			if receivedToken == "" {
				http.Error(w, "CSRF token missing", http.StatusForbidden)
				return
			}

			cookie, err := r.Cookie(cfg.CookieName)
			if err != nil {
				http.Error(w, "CSRF cookie missing", http.StatusForbidden)
				return
			}

			if !hmac.Equal([]byte(receivedToken), []byte(cookie.Value)) {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func parsedSameSite(cfg config.CSRFConfig) http.SameSite {
	switch cfg.SameSite {
	case "Lax":
		return http.SameSiteLaxMode
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
