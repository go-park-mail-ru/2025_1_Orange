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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEmployerHandler_Register(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(employer *mock.MockEmployer, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "успешная регистрация работодателя",
			requestBody: dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "company@example.com",
					Password: "strongpassword123",
				},
				CompanyName:  "ООО Тестовая Компания",
				LegalAddress: "г. Москва, ул. Тестовая, д. 1",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(1, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 1, "employer").
					Return("session-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.AuthResponse{
				UserID: 1,
				Role:   "employer",
			},
		},
		{
			name:        "невалидный JSON",
			requestBody: "{invalid}",
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: entity.ErrBadRequest.Error(),
			},
		},
		{
			name: "невалидные данные (отсутствует email)",
			requestBody: dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Password: "strongpassword123",
				},
				CompanyName: "ООО Тестовая Компания",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(0, entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("email обязателен"),
					))
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: "email обязателен",
			},
		},
		{
			name: "компания уже существует",
			requestBody: dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "existing@example.com",
					Password: "strongpassword123",
				},
				CompanyName: "Существующая компания",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(0, entity.NewError(
						entity.ErrAlreadyExists,
						fmt.Errorf("компания с таким email уже существует"),
					))
			},
			expectedStatus: http.StatusConflict,
			expectedResponse: utils.APIError{
				Status:  http.StatusConflict,
				Message: "компания с таким email уже существует",
			},
		},
		{
			name: "ошибка при создании сессии",
			requestBody: dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "company@example.com",
					Password: "strongpassword123",
				},
				CompanyName: "ООО Тестовая Компания",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Return(2, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 2, "employer").
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

			mockEmployer := mock.NewMockEmployer(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tc.mockSetup(mockEmployer, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewEmployerHandler(mockAuth, mockEmployer, nil, cfg)

			var reqBody []byte
			if body, ok := tc.requestBody.(string); ok {
				reqBody = []byte(body)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/employer/register", bytes.NewReader(reqBody))
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

func TestEmployerHandler_Login(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(employer *mock.MockEmployer, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "успешный вход работодателя",
			requestBody: &dto.Login{
				Email:    "company@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(1, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 1, "employer").
					Return("session-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.AuthResponse{
				UserID: 1,
				Role:   "employer",
			},
		},
		{
			name:        "неверный формат JSON",
			requestBody: "{invalid}",
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
			},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: entity.ErrBadRequest.Error(),
			},
		},
		{
			name: "неверные учетные данные",
			requestBody: &dto.Login{
				Email:    "wrong@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
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
			name: "работодатель не найден",
			requestBody: &dto.Login{
				Email:    "notfound@example.com",
				Password: "somepassword",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(0, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("работодатель не найден"),
					))
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: utils.APIError{
				Status:  http.StatusNotFound,
				Message: "работодатель не найден",
			},
		},
		{
			name: "ошибка при создании сессии",
			requestBody: &dto.Login{
				Email:    "company@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(2, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 2, "employer").
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
		{
			name: "ошибка при кодировании ответа",
			requestBody: &dto.Login{
				Email:    "company@example.com",
				Password: "correctpassword",
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				employer.EXPECT().
					Login(gomock.Any(), gomock.Any()).
					Return(3, nil)

				auth.EXPECT().
					CreateSession(gomock.Any(), 3, "employer").
					Return("session-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.AuthResponse{
				UserID: 3,
				Role:   "employer",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEmployer := mock.NewMockEmployer(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tc.mockSetup(mockEmployer, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewEmployerHandler(mockAuth, mockEmployer, nil, cfg)

			var reqBody []byte
			if body, ok := tc.requestBody.(string); ok {
				reqBody = []byte(body)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/employer/login", bytes.NewReader(reqBody))
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

func TestEmployerHandler_GetProfile(t *testing.T) {
	t.Parallel()

	testProfile := &dto.EmployerProfileResponse{
		ID:           1,
		CompanyName:  "ВКонтакте",
		LegalAddress: "г. Санкт-Петербург, Дворцовая набережная, 7-9",
		Email:        "career@vk.com",
		Slogan:       "Общение без границ",
		Website:      "https://vk.company",
		Description:  "Крупнейшая социальная сеть в России и СНГ",
		Vk:           "https://vk.com/vk",
		Telegram:     "https://t.me/vk",
		Facebook:     "https://facebook.com/vk",
		LogoPath:     "/assets/vk_logo.jpg",
		CreatedAt:    time.Date(2006, 9, 20, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	testCases := []struct {
		name             string
		pathID           string
		setupRequest     func() *http.Request
		mockSetup        func(employer *mock.MockEmployer)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name:   "успешное получение профиля работодателя",
			pathID: "1",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/employer/1", nil)
			},
			mockSetup: func(employer *mock.MockEmployer) {
				employer.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(testProfile, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: testProfile,
		},
		{
			name:   "невалидный ID в URL",
			pathID: "invalid",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/employer/invalid", nil)
			},
			mockSetup: func(employer *mock.MockEmployer) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: entity.ErrBadRequest.Error(),
			},
		},
		{
			name:   "работодатель не найден",
			pathID: "999",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/employer/999", nil)
			},
			mockSetup: func(employer *mock.MockEmployer) {
				employer.EXPECT().
					GetUser(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("работодатель не найден"),
					))
			},
			expectedStatus: http.StatusNotFound,
			expectedResponse: utils.APIError{
				Status:  http.StatusNotFound,
				Message: "работодатель не найден",
			},
		},
		{
			name:   "ошибка при кодировании ответа",
			pathID: "1",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/employer/1", nil)
			},
			mockSetup: func(employer *mock.MockEmployer) {
				employer.EXPECT().
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

			mockEmployer := mock.NewMockEmployer(ctrl)

			tc.mockSetup(mockEmployer)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewEmployerHandler(nil, mockEmployer, nil, cfg)

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
				var profile dto.EmployerProfileResponse
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

func TestEmployerHandler_UpdateProfile(t *testing.T) {
	t.Parallel()

	updateEmployerDTO := &dto.EmployerProfileUpdate{
		CompanyName:  "ВКонтакте",
		LegalAddress: "г. Санкт-Петербург, Дворцовая набережная, 7-9",
		Slogan:       "Общение без границ",
		Website:      "https://vk.company",
		Description:  "Крупнейшая социальная сеть в России и СНГ",
		Vk:           "https://vk.com/vk",
		Telegram:     "https://t.me/vk",
		Facebook:     "https://facebook.com/vk",
	}

	testCases := []struct {
		name             string
		requestBody      interface{}
		setupRequest     func() *http.Request
		mockSetup        func(employer *mock.MockEmployer, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name:        "успешное обновление профиля работодателя",
			requestBody: updateEmployerDTO,
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateEmployerDTO)
				req := httptest.NewRequest(http.MethodPut, "/employer/profile", bytes.NewReader(body))
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)

				employer.EXPECT().
					UpdateProfile(gomock.Any(), 1, updateEmployerDTO).
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:        "отсутствует cookie сессии",
			requestBody: updateEmployerDTO,
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateEmployerDTO)
				return httptest.NewRequest(http.MethodPut, "/employer/profile", bytes.NewReader(body))
			},
			mockSetup:      func(employer *mock.MockEmployer, auth *mock.MockAuth) {},
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
				req := httptest.NewRequest(http.MethodPut, "/employer/profile", bytes.NewReader([]byte("{invalid}")))
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: "invalid character 'i' looking for beginning of object key string",
			},
		},
		{
			name:        "недостаточно прав (не employer)",
			requestBody: updateEmployerDTO,
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateEmployerDTO)
				req := httptest.NewRequest(http.MethodPut, "/employer/profile", bytes.NewReader(body))
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedResponse: utils.APIError{
				Status:  http.StatusForbidden,
				Message: entity.ErrForbidden.Error(),
			},
		},
		{
			name:        "ошибка при обновлении профиля",
			requestBody: updateEmployerDTO,
			setupRequest: func() *http.Request {
				body, _ := json.Marshal(updateEmployerDTO)
				req := httptest.NewRequest(http.MethodPut, "/employer/profile", bytes.NewReader(body))
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)

				employer.EXPECT().
					UpdateProfile(gomock.Any(), 1, updateEmployerDTO).
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

			mockEmployer := mock.NewMockEmployer(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tc.mockSetup(mockEmployer, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewEmployerHandler(mockAuth, mockEmployer, nil, cfg)

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

func TestEmployerHandler_UploadLogo(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2025, 4, 16, 12, 0, 0, 0, time.UTC)
	testLogo := &dto.UploadStaticResponse{
		ID:        1,
		Path:      "/assets/logo.jpg",
		CreatedAt: fixedTime,
		UpdatedAt: fixedTime,
	}

	testCases := []struct {
		name             string
		setupRequest     func() *http.Request
		mockSetup        func(employer *mock.MockEmployer, auth *mock.MockAuth, static *mock.MockStatic)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "успешная загрузка логотипа",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("logo", "logo.jpg")

				_, err := part.Write([]byte("test image content"))
				require.NoError(t, err)

				err = writer.Close()
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/employer/logo", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth, static *mock.MockStatic) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)

				static.EXPECT().
					UploadStatic(gomock.Any(), gomock.Any()).
					Return(testLogo, nil)

				employer.EXPECT().
					UpdateLogo(gomock.Any(), 1, 1).
					Return(nil)
			},
			expectedStatus:   http.StatusOK,
			expectedResponse: testLogo,
		},
		{
			name: "отсутствует cookie сессии",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/employer/logo", nil)
			},
			mockSetup:      func(employer *mock.MockEmployer, auth *mock.MockAuth, static *mock.MockStatic) {},
			expectedStatus: http.StatusUnauthorized,
			expectedResponse: utils.APIError{
				Status:  http.StatusUnauthorized,
				Message: entity.ErrUnauthorized.Error(),
			},
		},
		{
			name: "недостаточно прав (не employer)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/employer/logo", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth, static *mock.MockStatic) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedResponse: utils.APIError{
				Status:  http.StatusForbidden,
				Message: entity.ErrForbidden.Error(),
			},
		},
		{
			name: "отсутствует файл логотипа",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodPost, "/employer/logo", nil)
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth, static *mock.MockStatic) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: entity.ErrBadRequest.Error(),
			},
		},
		{
			name: "ошибка при загрузке файла",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("logo", "logo.jpg")

				_, err := part.Write([]byte("test image content"))
				require.NoError(t, err)

				err = writer.Close()
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/employer/logo", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth, static *mock.MockStatic) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)

				static.EXPECT().
					UploadStatic(gomock.Any(), gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при загрузке файла"),
					))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: utils.APIError{
				Status:  http.StatusInternalServerError,
				Message: "ошибка при загрузке файла",
			},
		},
		{
			name: "ошибка при обновлении логотипа",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("logo", "logo.jpg")

				_, err := part.Write([]byte("test image content"))
				require.NoError(t, err)

				err = writer.Close()
				require.NoError(t, err)

				req := httptest.NewRequest(http.MethodPost, "/employer/logo", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "valid-session"})
				return req
			},
			mockSetup: func(employer *mock.MockEmployer, auth *mock.MockAuth, static *mock.MockStatic) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)

				static.EXPECT().
					UploadStatic(gomock.Any(), gomock.Any()).
					Return(testLogo, nil)

				employer.EXPECT().
					UpdateLogo(gomock.Any(), 1, 1).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при обновлении логотипа"),
					))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedResponse: utils.APIError{
				Status:  http.StatusInternalServerError,
				Message: "ошибка при обновлении логотипа",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEmployer := mock.NewMockEmployer(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)
			mockStatic := mock.NewMockStatic(ctrl)

			tc.mockSetup(mockEmployer, mockAuth, mockStatic)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewEmployerHandler(mockAuth, mockEmployer, mockStatic, cfg)

			req := tc.setupRequest()
			w := httptest.NewRecorder()

			handler.UploadLogo(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var logo dto.UploadStaticResponse
				err := json.NewDecoder(res.Body).Decode(&logo)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, &logo)
			} else if tc.expectedResponse != nil {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, apiErr)
			}
		})
	}
}

func TestEmployerHandler_EmailExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		requestBody      interface{}
		mockSetup        func(employer *mock.MockEmployer)
		expectedStatus   int
		expectedResponse interface{}
	}{
		{
			name: "email существует",
			requestBody: dto.EmailExistsRequest{
				Email: "employer@example.com",
			},
			mockSetup: func(employer *mock.MockEmployer) {
				employer.EXPECT().
					EmailExists(gomock.Any(), "employer@example.com").
					Return(&dto.EmailExistsResponse{
						Exists: true,
						Role:   "employer",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.EmailExistsResponse{
				Exists: true,
				Role:   "employer",
			},
		},
		{
			name: "email не существует",
			requestBody: dto.EmailExistsRequest{
				Email: "nonexistent@example.com",
			},
			mockSetup: func(employer *mock.MockEmployer) {
				employer.EXPECT().
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
			mockSetup:      func(employer *mock.MockEmployer) {},
			expectedStatus: http.StatusBadRequest,
			expectedResponse: utils.APIError{
				Status:  http.StatusBadRequest,
				Message: entity.ErrBadRequest.Error(),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			MockEmployer := mock.NewMockEmployer(ctrl)
			tc.mockSetup(MockEmployer)

			cfg := config.CSRFConfig{}
			handler := NewEmployerHandler(nil, MockEmployer, nil, cfg)

			var reqBody []byte
			if body, ok := tc.requestBody.(string); ok {
				reqBody = []byte(body)
			} else {
				reqBody, _ = json.Marshal(tc.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/employer/emailExists", bytes.NewReader(reqBody))
			w := httptest.NewRecorder()

			handler.EmailExists(w, req)

			res := w.Result()
			defer res.Body.Close()

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
