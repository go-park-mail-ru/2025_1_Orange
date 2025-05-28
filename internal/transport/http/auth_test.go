package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase/mock"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAuthHandler_IsAuth(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		setupRequest     func() *http.Request
		mockSetup        func(auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "успешная проверка аутентификации (applicant)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/auth/isAuth", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.AuthResponse{
				UserID: 1,
				Role:   "applicant",
			},
		},
		{
			name: "успешная проверка аутентификации (employer)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/auth/isAuth", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(2, "employer", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.AuthResponse{
				UserID: 2,
				Role:   "employer",
			},
		},
		{
			name: "отсутствует cookie сессии",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/auth/isAuth", nil)
			},
			mockSetup:      func(auth *mock.MockAuth) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: entity.ErrUnauthorized.Error(),
			},
		},
		{
			name: "недействительная сессия",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/auth/isAuth", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "invalid-session").
					Return(0, "", entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("сессия не найдена"),
					))
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: utils.APIError{
				Status:  http.StatusNotFound,
				Message: "сессия не найдена",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			tc.mockSetup(mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewAuthHandler(mockAuth, cfg)

			req := tc.setupRequest()
			w := httptest.NewRecorder()

			handler.IsAuth(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var authResponse dto.AuthResponse
				err := json.NewDecoder(res.Body).Decode(&authResponse)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, authResponse)
			} else {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		setupRequest     func() *http.Request
		mockSetup        func(auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
		checkCookies     bool
	}{
		{
			name: "успешный выход",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					Logout(gomock.Any(), "valid-session").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkCookies:   true,
		},
		{
			name: "выход без cookie сессии",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
			},
			mockSetup:      func(auth *mock.MockAuth) {},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ошибка при выходе из системы",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					Logout(gomock.Any(), "invalid-session").
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при выходе из системы"),
					))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: utils.APIError{
				Status:  http.StatusInternalServerError,
				Message: "ошибка при выходе из системы",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			tc.mockSetup(mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewAuthHandler(mockAuth, cfg)

			req := tc.setupRequest()
			w := httptest.NewRecorder()

			handler.Logout(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if tc.checkCookies {
				cookies := res.Cookies()
				var sessionCookie, csrfCookie *http.Cookie
				for _, cookie := range cookies {
					switch cookie.Name {
					case "session_id":
						sessionCookie = cookie
					case cfg.CookieName:
						csrfCookie = cookie
					}
				}

				require.NotNil(t, sessionCookie)
				require.Equal(t, "", sessionCookie.Value)
				require.True(t, sessionCookie.Expires.Before(time.Now()))

				require.NotNil(t, csrfCookie)
				require.NotEmpty(t, csrfCookie.Value)
			}

			if tc.expectedResponse != nil {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}

func TestAuthHandler_LogoutAll(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		setupRequest     func() *http.Request
		mockSetup        func(auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
		checkCookies     bool
	}{
		{
			name: "успешный выход из всех устройств (applicant)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/auth/logoutAll", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				auth.EXPECT().
					LogoutAll(gomock.Any(), 1, "applicant").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkCookies:   true,
		},
		{
			name: "успешный выход из всех устройств (employer)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/auth/logoutAll", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(2, "employer", nil)
				auth.EXPECT().
					LogoutAll(gomock.Any(), 2, "employer").
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			checkCookies:   true,
		},
		{
			name: "выход из всех устройств без cookie сессии",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/auth/logoutAll", nil)
			},
			mockSetup:      func(auth *mock.MockAuth) {},
			expectedStatus: http.StatusOK,
		},
		{
			name: "ошибка при получении сессии",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/auth/logoutAll", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "invalid-session").
					Return(0, "", entity.NewError(
						entity.ErrUnauthorized,
						fmt.Errorf("сессия не найдена"),
					))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: "сессия не найдена",
			},
		},
		{
			name: "ошибка при выходе из всех устройств",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/auth/logoutAll", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				auth.EXPECT().
					LogoutAll(gomock.Any(), 1, "applicant").
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка базы данных"),
					))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: utils.APIError{
				Status:  http.StatusInternalServerError,
				Message: "ошибка базы данных",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			tc.mockSetup(mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewAuthHandler(mockAuth, cfg)

			req := tc.setupRequest()
			w := httptest.NewRecorder()

			handler.LogoutAll(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if tc.checkCookies {
				cookies := res.Cookies()
				var sessionCookie, csrfCookie *http.Cookie
				for _, cookie := range cookies {
					switch cookie.Name {
					case "session_id":
						sessionCookie = cookie
					case cfg.CookieName:
						csrfCookie = cookie
					}
				}

				require.NotNil(t, sessionCookie)
				require.Equal(t, "", sessionCookie.Value)
				require.True(t, sessionCookie.Expires.Before(time.Now()))

				require.NotNil(t, csrfCookie)
				require.NotEmpty(t, csrfCookie.Value)
			}

			if tc.expectedResponse != nil {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}
