package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase/mock"
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestApplicantHandler_Register(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		requestBody      *dto.ApplicantRegister
		mockSetup        func(applicant *mock.MockApplicant, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "успешная регистрация",
			requestBody: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "test@example.com",
					Password: "strongpassword",
				},
				FirstName: "Имя",
				LastName:  "Фамилия",
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				applicant.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(1, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 1, "applicant").
					Return("session-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.AuthResponse{
				UserID: 1,
				Role:   "applicant",
			},
		},
		{
			name:        "неверный формат JSON",
			requestBody: nil,
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: "невалидный json: parse error: syntax error near offset 1 of '{invalid}'",
			},
		},
		{
			name: "невалидные данные",
			requestBody: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "invalid_email",
					Password: "strong!Password",
				},
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				applicant.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(0, entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("неверный формат данных"),
					))
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: "неверный формат данных",
			},
		},
		{
			name: "пользователь уже существует",
			requestBody: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "existing@example.com",
					Password: "somepassword",
				},
				FirstName: "Имя",
				LastName:  "Фамилия",
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				applicant.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(0, entity.NewError(
						entity.ErrAlreadyExists,
						fmt.Errorf("такой пользователь уже существует"),
					))
			},
			expectedStatus: http.StatusConflict,
			expectedResponse: utils.APIError{
				Status:  http.StatusConflict,
				Message: "такой пользователь уже существует",
			},
		},
		{
			name: "ошибка сессии",
			requestBody: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "user@example.com",
					Password: "somepassword",
				},
				FirstName: "Имя",
				LastName:  "Фамилия",
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				applicant.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(2, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 2, "applicant").
					Return("", entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось создать сессию"),
					))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: utils.APIError{
				Status:  http.StatusInternalServerError,
				Message: "не удалось создать сессию",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApplicant := mock.NewMockApplicant(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tc.mockSetup(mockApplicant, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewApplicantHandler(mockAuth, mockApplicant, cfg)

			var reqBody []byte
			if tc.requestBody != nil {
				reqBody, _ = json.Marshal(tc.requestBody)
			} else {
				reqBody = []byte("{invalid}")
			}
			req := httptest.NewRequest(http.MethodPost, "/applicant/register", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var success dto.AuthResponse
				err := json.NewDecoder(res.Body).Decode(&success)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, success)
			} else {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}

func TestApplicantHandler_Login(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		requestBody      *dto.Login
		mockSetup        func(applicant *mock.MockApplicant, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "успешный вход",
			requestBody: &dto.Login{
				Email:    "test@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				applicant.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(1, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 1, "applicant").
					Return("session-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.AuthResponse{
				UserID: 1,
				Role:   "applicant",
			},
		},
		{
			name:        "неверный формат JSON",
			requestBody: nil,
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: "невалидный json: parse error: syntax error near offset 1 of '{invalid}'",
			},
		},
		{
			name: "неверные учетные данные",
			requestBody: &dto.Login{
				Email:    "wrong@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				applicant.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(0, entity.NewError(
						entity.ErrUnauthorized,
						fmt.Errorf("неверные учетные данные"),
					))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: "неверные учетные данные",
			},
		},
		{
			name: "ошибка при создании сессии",
			requestBody: &dto.Login{
				Email:    "test@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				applicant.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(2, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 2, "applicant").
					Return("", entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при создании сессии"),
					))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: utils.APIError{
				Status:  http.StatusInternalServerError,
				Message: "ошибка при создании сессии",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApplicant := mock.NewMockApplicant(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tc.mockSetup(mockApplicant, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewApplicantHandler(mockAuth, mockApplicant, cfg)

			var reqBody []byte
			if tc.requestBody != nil {
				reqBody, _ = json.Marshal(tc.requestBody)
			} else {
				reqBody = []byte("{invalid}")
			}

			req := httptest.NewRequest(http.MethodPost, "/applicant/login", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var success dto.AuthResponse
				err := json.NewDecoder(res.Body).Decode(&success)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, success)
			} else {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}

func TestApplicantHandler_GetProfile(t *testing.T) {
	t.Parallel()

	testProfile := &dto.ApplicantProfileResponse{
		ID:         1,
		FirstName:  "Иван",
		LastName:   "Иванов",
		MiddleName: "Иванович",
		City:       "Москва",
		BirthDate:  time.Date(1990, 4, 2, 0, 0, 0, 0, time.UTC),
		Sex:        "M",
		Email:      "ivan@example.com",
		Status:     "actively_searching",
		Quote:      "Когда волк молчит - лучше не перебивать",
		Vk:         "https://vk.com/ivanich",
		Telegram:   "https://t.me/ivanich",
		Facebook:   "https://facebook.com/ivanich",
		AvatarPath: "/assets/1.jpg",
		CreatedAt:  time.Date(1990, 4, 2, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(1990, 4, 2, 0, 0, 0, 0, time.UTC),
	}

	testCases := []struct {
		name             string
		pathID           string
		setupRequest     func() *http.Request
		mockSetup        func(applicant *mock.MockApplicant, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name:   "успешное получение профиля",
			pathID: "1",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/applicant/1", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)

				applicant.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(testProfile, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: testProfile,
		},
		{
			name:   "отсутствует cookie сессии",
			pathID: "1",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/applicant/1", nil)
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: entity.ErrUnauthorized.Error(),
			},
		},
		{
			name:   "невалидный ID в URL",
			pathID: "invalid",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/applicant/invalid", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: entity.ErrBadRequest.Error(),
			},
		},
		{
			name:   "ошибка при проверке сессии",
			pathID: "1",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/applicant/1", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "invalid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
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
			name:   "профиль не найден",
			pathID: "999",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/applicant/999", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)

				applicant.EXPECT().
					GetUser(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("профиль не найден"),
					))
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: utils.APIError{
				Status:  http.StatusNotFound,
				Message: "профиль не найден",
			},
		},
		{
			name:   "ошибка при кодировании ответа",
			pathID: "1",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/applicant/1", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)

				applicant.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(testProfile, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: testProfile,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApplicant := mock.NewMockApplicant(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tc.mockSetup(mockApplicant, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewApplicantHandler(mockAuth, mockApplicant, cfg)

			req := tc.setupRequest()
			req.SetPathValue("id", tc.pathID)
			w := httptest.NewRecorder()

			handler.GetProfile(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var profile dto.ApplicantProfileResponse
				err := json.NewDecoder(res.Body).Decode(&profile)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, &profile)
			} else {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}

func TestApplicantHandler_UpdateProfile(t *testing.T) {
	t.Parallel()

	updateApplicantDTO := &dto.ApplicantProfileUpdate{
		FirstName:  "Иван",
		LastName:   "Иванов",
		MiddleName: "Иванович",
		City:       "Москва",
		BirthDate:  time.Now().AddDate(-20, 0, 0),
		Sex:        "M",
		Status:     "actively_searching",
		Quote:      "Когда волк молчит - лучше не перебивать",
		Vk:         "https://vk.com/ivanich",
		Telegram:   "https://t.me/ivanich",
		Facebook:   "https://facebook.com/ivanich",
	}

	testCases := []struct {
		name             string
		requestBody      interface{}
		setupRequest     func() *http.Request
		mockSetup        func(applicant *mock.MockApplicant, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name:        "отсутствует cookie сессии",
			requestBody: updateApplicantDTO,
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateApplicantDTO)
				return httptest.NewRequest(http.MethodPut, "/applicant/profile", bytes.NewReader(body))
			},
			mockSetup:      func(applicant *mock.MockApplicant, auth *mock.MockAuth) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: entity.ErrUnauthorized.Error(),
			},
		},
		{
			name:        "невалидный JSON",
			requestBody: "{invalid}",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPut, "/applicant/profile", bytes.NewReader([]byte("{invalid}")))
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: "невалидный json: parse error: syntax error near offset 1 of '{invalid}'",
			},
		},
		{
			name:        "недостаточно прав (не applicant)",
			requestBody: updateApplicantDTO,
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateApplicantDTO)
				req := httptest.NewRequest(http.MethodPut, "/applicant/profile", bytes.NewReader(body))
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedResponse: utils.APIError{
				Status:  http.StatusForbidden,
				Message: entity.ErrForbidden.Error(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApplicant := mock.NewMockApplicant(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tc.mockSetup(mockApplicant, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewApplicantHandler(mockAuth, mockApplicant, cfg)

			req := tc.setupRequest()
			w := httptest.NewRecorder()

			handler.UpdateProfile(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if tc.expectedStatus != http.StatusNoContent {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			} else {
				require.Empty(t, w.Body.Bytes())
			}
		})
	}
}

func TestApplicantHandler_UploadAvatar(t *testing.T) {
	t.Parallel()

	testAvatar := &dto.UploadStaticResponse{
		ID:   1,
		Path: "/assets/avatar.jpg",
	}

	testCases := []struct {
		name             string
		setupRequest     func() *http.Request
		mockSetup        func(applicant *mock.MockApplicant, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "успешная загрузка аватара",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("avatar", "avatar.jpg")

				_, err := part.Write([]byte("test image content"))
				require.NoError(t, err)

				err = writer.Close()
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/applicant/avatar", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)

				applicant.EXPECT().
					UpdateAvatar(gomock.Any(), 1, gomock.Any()).
					Return(testAvatar, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: testAvatar,
		},
		{
			name: "отсутствует cookie сессии",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/applicant/avatar", nil)
			},
			mockSetup:      func(applicant *mock.MockApplicant, auth *mock.MockAuth) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: entity.ErrUnauthorized.Error(),
			},
		},
		{
			name: "недостаточно прав (не applicant)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/applicant/avatar", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedResponse: utils.APIError{
				Status:  http.StatusForbidden,
				Message: entity.ErrForbidden.Error(),
			},
		},
		{
			name: "отсутствует файл аватара",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/applicant/avatar", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: entity.ErrBadRequest.Error(),
			},
		},
		{
			name: "ошибка при обновлении аватара",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("avatar", "avatar.jpg")

				_, err := part.Write([]byte("test image content"))
				require.NoError(t, err)

				err = writer.Close()
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/applicant/avatar", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)

				applicant.EXPECT().
					UpdateAvatar(gomock.Any(), 1, gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при обновлении аватара"),
					))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: utils.APIError{
				Status:  http.StatusInternalServerError,
				Message: "ошибка при обновлении аватара",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApplicant := mock.NewMockApplicant(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tc.mockSetup(mockApplicant, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewApplicantHandler(mockAuth, mockApplicant, cfg)

			req := tc.setupRequest()
			w := httptest.NewRecorder()

			handler.UploadAvatar(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var avatar dto.UploadStaticResponse
				err := json.NewDecoder(res.Body).Decode(&avatar)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, &avatar)
			} else if tc.expectedResponse != nil {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}

func TestApplicantHandler_EmailExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(applicant *mock.MockApplicant)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "email существует",
			requestBody: dto.EmailExistsRequest{
				Email: "applicant@example.com",
			},
			mockSetup: func(applicant *mock.MockApplicant) {
				applicant.EXPECT().
					EmailExists(gomock.Any(), "applicant@example.com").
					Return(&dto.EmailExistsResponse{
						Exists: true,
						Role:   "applicant",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.EmailExistsResponse{
				Exists: true,
				Role:   "applicant",
			},
		},
		{
			name: "email не существует",
			requestBody: dto.EmailExistsRequest{
				Email: "nonexistent@example.com",
			},
			mockSetup: func(applicant *mock.MockApplicant) {
				applicant.EXPECT().
					EmailExists(gomock.Any(), "nonexistent@example.com").
					Return(nil, entity.NewError(entity.ErrNotFound, fmt.Errorf("почта не найдена")))
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: utils.APIError{
				Status:  http.StatusNotFound,
				Message: "почта не найдена",
			},
		},
		{
			name:           "невалидный JSON",
			requestBody:    "{invalid}",
			mockSetup:      func(applicant *mock.MockApplicant) {},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: "невалидный json: parse error: syntax error near offset 1 of '{invalid}'",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			MockApplicant := mock.NewMockApplicant(ctrl)
			tc.mockSetup(MockApplicant)

			cfg := config.CSRFConfig{}
			handler := NewApplicantHandler(nil, MockApplicant, cfg)

			var reqBody []byte
			if body, ok := tc.requestBody.(string); ok {
				reqBody = []byte(body)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/applicant/emailExists", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			handler.EmailExists(w, req)

			res := w.Result()

			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var response dto.EmailExistsResponse
				err := json.NewDecoder(res.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, response)
			} else {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}
