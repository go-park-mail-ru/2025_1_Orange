package middleware

import (
	"ResuMatch/internal/config"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
)

type CSRF struct {
	secret     []byte
	cookieName string
	cookieCfg  config.CookiesConfig
}

func NewCSRF(secret, cookieName string, cookieCfg config.CookiesConfig) *CSRF {
	return &CSRF{
		secret:     []byte(secret),
		cookieName: cookieName,
		cookieCfg:  cookieCfg,
	}
}

func (c *CSRF) generateToken(r *http.Request) string {
	h := hmac.New(sha256.New, c.secret)
	// Соль в виде ip адреса
	h.Write([]byte(r.RemoteAddr))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (c *CSRF) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Исключаем безопасные методы
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			token := c.generateToken(r)
			http.SetCookie(w, &http.Cookie{
				Name:  c.cookieName,
				Value: token,
				// устанавливаем куку для всех маршрутов домена
				Path:     "/",
				Secure:   c.cookieCfg.Secure,
				HttpOnly: c.cookieCfg.HTTPOnly,
				SameSite: c.parseSameSite(),
			})
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем токен для небезопасных методов
		receivedToken := r.Header.Get("X-CSRF-Token")
		if receivedToken == "" {
			http.Error(w, "CSRF token missing", http.StatusForbidden)
			return
		}

		cookie, err := r.Cookie(c.cookieName)
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

func (c *CSRF) parseSameSite() http.SameSite {
	switch c.cookieCfg.SameSite {
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
