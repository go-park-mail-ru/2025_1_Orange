package middleware

import (
	"ResuMatch/internal/config"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	cfg := config.CSRFConfig{Secret: "test-secret"}
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	t.Run("With session ID", func(t *testing.T) {
		t.Parallel()
		sessionID := "test-session"
		token := generateToken(req, sessionID, cfg)
		require.NotEmpty(t, token)
		require.Len(t, token, 44) // Base64 encoded SHA256
	})

	t.Run("Without session ID", func(t *testing.T) {
		t.Parallel()
		token := generateToken(req, "", cfg)
		require.NotEmpty(t, token)
	})
}

func TestSetCSRFToken(t *testing.T) {
	t.Parallel()

	cfg := config.CSRFConfig{
		CookieName: "csrf_token",
		Secret:     "test-secret",
		Secure:     true,
		HttpOnly:   true,
		Lifetime:   time.Hour,
		SameSite:   "Lax",
	}

	t.Run("With session cookie", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.AddCookie(&http.Cookie{Name: "session_id", Value: "test-session"})

		SetCSRFToken(w, r, cfg)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		require.Equal(t, cfg.CookieName, cookies[0].Name)
		require.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)
		require.Equal(t, cfg.Secure, cookies[0].Secure)
		require.Equal(t, cfg.HttpOnly, cookies[0].HttpOnly)

		tokenHeader := w.Header().Get("X-CSRF-Token")
		require.NotEmpty(t, tokenHeader)
		require.Equal(t, cookies[0].Value, tokenHeader)
	})

	t.Run("Without session cookie", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		SetCSRFToken(w, r, cfg)

		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		require.NotEmpty(t, cookies[0].Value)
	})
}

func TestCSRFMiddleware(t *testing.T) {
	t.Parallel()

	cfg := config.CSRFConfig{
		CookieName: "csrf_token",
		Secret:     "test-secret",
		Lifetime:   time.Hour,
	}

	t.Run("Safe methods", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/api/test"},
			{http.MethodHead, "/api/test"},
			{http.MethodOptions, "/api/test"},
			{http.MethodGet, "/api/v1/static/file.txt"},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.method+" "+tc.path, func(t *testing.T) {
				t.Parallel()
				w := httptest.NewRecorder()
				r := httptest.NewRequest(tc.method, tc.path, nil)

				middleware := CSRFMiddleware(cfg)
				handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				handler.ServeHTTP(w, r)

				if strings.HasPrefix(tc.path, "/api/v1/static/") {
					require.Empty(t, w.Header().Get("X-CSRF-Token"))
				} else {
					require.NotEmpty(t, w.Header().Get("X-CSRF-Token"))
				}
				require.Equal(t, http.StatusOK, w.Code)
			})
		}
	})

	t.Run("Unsafe methods - valid token", func(t *testing.T) {
		t.Parallel()
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/test", nil)

		// Сначала получаем токен
		token := generateToken(r, "", cfg)
		r.Header.Set("X-CSRF-Token", token)
		r.AddCookie(&http.Cookie{Name: cfg.CookieName, Value: token})

		middleware := CSRFMiddleware(cfg)
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Unsafe methods - invalid cases", func(t *testing.T) {
		t.Parallel()
		testCases := []struct {
			name        string
			tokenHeader string
			tokenCookie string
			expected    int
		}{
			{
				name:     "Missing token header",
				expected: http.StatusForbidden,
			},
			{
				name:        "Missing cookie",
				tokenHeader: "token",
				expected:    http.StatusForbidden,
			},
			{
				name:        "Token mismatch",
				tokenHeader: "token1",
				tokenCookie: "token2",
				expected:    http.StatusForbidden,
			},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodPost, "/api/test", nil)

				if tc.tokenHeader != "" {
					r.Header.Set("X-CSRF-Token", tc.tokenHeader)
				}
				if tc.tokenCookie != "" {
					r.AddCookie(&http.Cookie{Name: cfg.CookieName, Value: tc.tokenCookie})
				}

				middleware := CSRFMiddleware(cfg)
				handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))

				handler.ServeHTTP(w, r)
				require.Equal(t, tc.expected, w.Code)
			})
		}
	})
}

func TestParsedSameSite(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected http.SameSite
	}{
		{"Lax", http.SameSiteLaxMode},
		{"Strict", http.SameSiteStrictMode},
		{"None", http.SameSiteNoneMode},
		{"Invalid", http.SameSiteDefaultMode},
		{"", http.SameSiteDefaultMode},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			cfg := config.CSRFConfig{SameSite: tc.input}
			result := parsedSameSite(cfg)
			require.Equal(t, tc.expected, result)
		})
	}
}
