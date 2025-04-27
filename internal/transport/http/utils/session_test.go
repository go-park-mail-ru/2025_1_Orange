package utils

import (
	// "ResuMatch/internal/usecase"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/usecase/mock"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestClearTokenCookies(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()

	ClearTokenCookies(w)

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 2)

	// Проверяем session_id cookie
	sessionCookie := cookies[0]
	require.Equal(t, "session_id", sessionCookie.Name)
	require.Equal(t, "", sessionCookie.Value)
	require.True(t, sessionCookie.Expires.Before(time.Now()))
	require.Equal(t, -1, sessionCookie.MaxAge)
	require.Equal(t, "/", sessionCookie.Path)
	require.True(t, sessionCookie.HttpOnly)
	require.Equal(t, http.SameSiteStrictMode, sessionCookie.SameSite)

	// Проверяем csrf_token cookie
	csrfCookie := cookies[1]
	require.Equal(t, "csrf_token", csrfCookie.Name)
	require.Equal(t, "", csrfCookie.Value)
	require.True(t, csrfCookie.Expires.Before(time.Now()))
	require.Equal(t, -1, csrfCookie.MaxAge)
	require.Equal(t, "/", csrfCookie.Path)
	require.True(t, csrfCookie.HttpOnly)
	require.Equal(t, http.SameSiteStrictMode, csrfCookie.SameSite)
}

func TestCreateSession_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authMock := mock.NewMockAuth(ctrl)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Ожидаем вызов CreateSession с любыми аргументами
	authMock.EXPECT().
		CreateSession(gomock.Any(), 123, "applicant").
		Return("session-token", nil)

	err := CreateSession(w, r, authMock, 123, "applicant")
	require.NoError(t, err)

	// Проверяем, что cookie установлен правильно
	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	sessionCookie := cookies[0]
	require.Equal(t, "session_id", sessionCookie.Name)
	require.Equal(t, "session-token", sessionCookie.Value)
	require.True(t, sessionCookie.Expires.After(time.Now()))
	require.Equal(t, "/", sessionCookie.Path)
	require.True(t, sessionCookie.HttpOnly)
	require.Equal(t, http.SameSiteStrictMode, sessionCookie.SameSite)
}

func TestCreateSession_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	authMock := mock.NewMockAuth(ctrl)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Ожидаем вызов CreateSession, который вернет ошибку
	expectedErr := entity.NewError(
		entity.ErrInternal,
		fmt.Errorf("не удалось получить сессию для пользователя с id=123, role=applicant"),
	)

	authMock.EXPECT().
		CreateSession(gomock.Any(), 123, "applicant").
		Return("", expectedErr)

	err := CreateSession(w, r, authMock, 123, "applicant")
	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func TestSetSession(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		value    string
		expires  time.Time
		expected http.Cookie
	}{
		{
			name:    "Normal session",
			value:   "session-token",
			expires: time.Now().Add(24 * time.Hour),
			expected: http.Cookie{
				Name:     "session_id",
				Value:    "session-token",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   false,
			},
		},
		{
			name:    "Empty session",
			value:   "",
			expires: time.Now().Add(-24 * time.Hour),
			expected: http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
				Secure:   false,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()

			SetSession(w, tc.value, tc.expires)

			cookies := w.Result().Cookies()
			require.Len(t, cookies, 1)

			actualCookie := cookies[0]
			require.Equal(t, tc.expected.Name, actualCookie.Name)
			require.Equal(t, tc.expected.Value, actualCookie.Value)
			require.Equal(t, tc.expected.Path, actualCookie.Path)
			require.Equal(t, tc.expected.HttpOnly, actualCookie.HttpOnly)
			require.Equal(t, tc.expected.SameSite, actualCookie.SameSite)
			require.Equal(t, tc.expected.Secure, actualCookie.Secure)

			if tc.value == "" {
				require.True(t, actualCookie.Expires.Before(time.Now()))
			} else {
				require.True(t, actualCookie.Expires.After(time.Now()))
			}
		})
	}
}
