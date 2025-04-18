package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"

	// "ResuMatch/internal/usecase"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase/mock"
	"bytes"

	// "context"
	"encoding/json"
	// "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResumeHandler_CreateResume(t *testing.T) {
	t.Parallel()

	// Определяем тестовые случаи
	testCases := []struct {
		name           string
		setupMocks     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		requestBody    interface{}
		cookie         *http.Cookie
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").Return(1, "applicant", nil)
				resume.EXPECT().Create(
					gomock.Any(),
					1,
					gomock.Any(),
				).Return(&dto.ResumeResponse{ID: 1}, nil)
			},
			requestBody: dto.CreateResumeRequest{
				AboutMe:                   "About me",
				Specialization:            "Backend",
				Education:                 entity.Higher,
				EducationalInstitution:    "University",
				GraduationYear:            "2020",
				Skills:                    []string{"Go", "SQL"},
				AdditionalSpecializations: []string{"DevOps"},
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Company",
						Position:     "Developer",
						Duties:       "Development",
						Achievements: "Achievements",
						StartDate:    "2020-01-01",
						EndDate:      "2022-01-01",
					},
				},
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "valid-session"},
			expectedStatus: http.StatusCreated,
			expectedBody:   &dto.ResumeResponse{ID: 1},
		},
		{
			name:           "Unauthorized - no cookie",
			setupMocks:     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			requestBody:    dto.CreateResumeRequest{},
			cookie:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
		},
		{
			name: "Unauthorized - invalid session",
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid-session").
					Return(0, "", entity.ErrUnauthorized)
			},
			requestBody:    dto.CreateResumeRequest{},
			cookie:         &http.Cookie{Name: "session_id", Value: "invalid-session"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.ToAPIError(entity.ErrInternal),
		},
		{
			name: "Forbidden - not applicant",
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").Return(1, "employer", nil)
			},
			requestBody:    dto.CreateResumeRequest{},
			cookie:         &http.Cookie{Name: "session_id", Value: "valid-session"},
			expectedStatus: http.StatusForbidden,
			expectedBody:   utils.APIError{Status: http.StatusForbidden, Message: entity.ErrForbidden.Error()},
		},
		{
			name: "Bad request - invalid body",
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").Return(1, "applicant", nil)
			},
			requestBody:    "invalid body",
			cookie:         &http.Cookie{Name: "session_id", Value: "valid-session"},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
		},
		{
			name: "Internal server error - resume creation failed",
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").Return(1, "applicant", nil)
				resume.EXPECT().Create(
					gomock.Any(),
					1,
					gomock.Any(),
				).Return(nil, entity.ErrInternal)
			},
			requestBody: dto.CreateResumeRequest{
				AboutMe:        "About me",
				Specialization: "Backend",
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "valid-session"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.APIError{Status: http.StatusInternalServerError, Message: entity.ErrInternal.Error()},
		},
	}

	// Запускаем тесты
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Создаем контроллер моков
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Создаем моки
			authMock := mock.NewMockAuth(ctrl)
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			// Настраиваем моки
			tc.setupMocks(authMock, resumeMock)

			// Создаем хэндлер с моками
			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

			// Создаем тестовый запрос
			var reqBody []byte
			switch body := tc.requestBody.(type) {
			case string:
				reqBody = []byte(body)
			default:
				var err error
				reqBody, err = json.Marshal(body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/resume/create", bytes.NewReader(reqBody))
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			// Создаем recorder для записи ответа
			rec := httptest.NewRecorder()

			// Вызываем хэндлер
			handler.CreateResume(rec, req)

			// Проверяем статус код
			require.Equal(t, tc.expectedStatus, rec.Code)

			// Проверяем тело ответа
			if tc.expectedStatus == http.StatusCreated {
				var resp dto.ResumeResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, &resp)
			} else {
				var resp utils.APIError
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, resp)
			}
		})
	}
}

func TestResumeHandler_GetResume(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		url            string // полный URL пути
		setupMocks     func(resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			url:  "/resume/123",
			setupMocks: func(resume *mock.MockResumeUsecase) {
				resume.EXPECT().GetByID(gomock.Any(), 123).
					Return(&dto.ResumeResponse{
						ID:             123,
						Specialization: "Backend",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &dto.ResumeResponse{
				ID:             123,
				Specialization: "Backend",
			},
		},
		{
			name: "Bad request - invalid resume ID",
			url:  "/resume/invalid",
			setupMocks: func(resume *mock.MockResumeUsecase) {
				// Мок не ожидает вызовов, так как ошибка будет до вызова usecase
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
		},
		{
			name: "Internal server error",
			url:  "/resume/123",
			setupMocks: func(resume *mock.MockResumeUsecase) {
				resume.EXPECT().GetByID(gomock.Any(), 123).
					Return(nil, entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.ToAPIError(entity.ErrInternal),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			tc.setupMocks(resumeMock)

			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

			// Используем http.NewRequest вместо httptest.NewRequest для правильного парсинга пути
			req, err := http.NewRequest(http.MethodGet, tc.url, nil)
			require.NoError(t, err)

			// Создаем router и регистрируем хэндлер
			router := http.NewServeMux()
			handler.Configure(router)

			rec := httptest.NewRecorder()

			// Вызываем router напрямую
			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp dto.ResumeResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, &resp)
			} else {
				var resp utils.APIError
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)

				expectedErr, ok := tc.expectedBody.(utils.APIError)
				require.True(t, ok)
				require.Equal(t, expectedErr.Status, resp.Status)
				require.Equal(t, expectedErr.Message, resp.Message)
			}
		})
	}
}

func TestResumeHandler_UpdateResume(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		url            string
		requestBody    interface{}
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success",
			url:  "/resume/123",
			requestBody: dto.UpdateResumeRequest{
				AboutMe:        "Updated about me",
				Specialization: "Updated specialization",
			},
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				resume.EXPECT().Update(
					gomock.Any(),
					123,
					1,
					gomock.Any(),
				).Return(&dto.ResumeResponse{ID: 123}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &dto.ResumeResponse{ID: 123},
		},
		{
			name:           "Unauthorized - no cookie",
			url:            "/resume/123",
			requestBody:    dto.UpdateResumeRequest{},
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
		},
		{
			name:        "Unauthorized - invalid session",
			url:         "/resume/123",
			requestBody: dto.UpdateResumeRequest{},
			cookie:      &http.Cookie{Name: "session_id", Value: "invalid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid-session").
					Return(0, "", entity.ErrUnauthorized)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.ToAPIError(entity.ErrInternal),
		},
		{
			name:        "Forbidden - not applicant",
			url:         "/resume/123",
			requestBody: dto.UpdateResumeRequest{},
			cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   utils.APIError{Status: http.StatusForbidden, Message: entity.ErrForbidden.Error()},
		},
		{
			name:        "Bad request - invalid body",
			url:         "/resume/123",
			requestBody: "invalid body",
			cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
		},
		{
			name: "Internal server error",
			url:  "/resume/123",
			requestBody: dto.UpdateResumeRequest{
				AboutMe: "Updated about me",
			},
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				resume.EXPECT().Update(
					gomock.Any(),
					123,
					1,
					gomock.Any(),
				).Return(nil, entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.ToAPIError(entity.ErrInternal),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			tc.setupMocks(authMock, resumeMock)

			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

			// Подготовка тела запроса
			var reqBody []byte
			switch body := tc.requestBody.(type) {
			case string:
				reqBody = []byte(body)
			default:
				var err error
				reqBody, err = json.Marshal(body)
				require.NoError(t, err)
			}

			req, err := http.NewRequest(http.MethodPut, tc.url, bytes.NewReader(reqBody))
			require.NoError(t, err)

			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			// Настройка роутера
			router := http.NewServeMux()
			handler.Configure(router)

			rec := httptest.NewRecorder()

			// Выполнение запроса
			router.ServeHTTP(rec, req)

			// Проверки
			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp dto.ResumeResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, &resp)
			} else {
				var resp utils.APIError
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)

				expectedErr, ok := tc.expectedBody.(utils.APIError)
				require.True(t, ok)
				require.Equal(t, expectedErr.Status, resp.Status)
				require.Equal(t, expectedErr.Message, resp.Message)
			}
		})
	}
}

func TestResumeHandler_DeleteResume(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		url            string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "Success",
			url:    "/resume/123",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				resume.EXPECT().Delete(gomock.Any(), 123, 1).
					Return(&dto.DeleteResumeResponse{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &dto.DeleteResumeResponse{Success: true},
		},
		{
			name:           "Unauthorized - no cookie",
			url:            "/resume/123",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
		},
		{
			name:   "Unauthorized - invalid session",
			url:    "/resume/123",
			cookie: &http.Cookie{Name: "session_id", Value: "invalid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid-session").
					Return(0, "", entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.ToAPIError(entity.ErrInternal),
		},
		{
			name:   "Forbidden - not applicant",
			url:    "/resume/123",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   utils.APIError{Status: http.StatusForbidden, Message: entity.ErrForbidden.Error()},
		},
		{
			name:   "Bad request - invalid resume ID",
			url:    "/resume/invalid",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				// Удаляем ожидание Times(0), так как хэндлер сначала проверяет куки
				// и только потом парсит ID
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				resume.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
		},
		{
			name:   "Internal server error",
			url:    "/resume/123",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				resume.EXPECT().Delete(gomock.Any(), 123, 1).
					Return(nil, entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.ToAPIError(entity.ErrInternal),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			tc.setupMocks(authMock, resumeMock)

			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

			req, err := http.NewRequest(http.MethodDelete, tc.url, nil)
			require.NoError(t, err)

			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			router := http.NewServeMux()
			handler.Configure(router)

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp dto.DeleteResumeResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, &resp)
			} else {
				var resp utils.APIError
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)

				expectedErr, ok := tc.expectedBody.(utils.APIError)
				require.True(t, ok)
				require.Equal(t, expectedErr.Status, resp.Status)
				require.Equal(t, expectedErr.Message, resp.Message)
			}
		})
	}
}

func TestResumeHandler_GetAllResumes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "Success - applicant role",
			cookie: &http.Cookie{Name: "session_id", Value: "applicant-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "applicant-session").
					Return(1, "applicant", nil)
				resume.EXPECT().GetAllResumesByApplicantID(gomock.Any(), 1).
					Return([]dto.ResumeShortResponse{
						{ID: 1, ApplicantID: 1},
						{ID: 2, ApplicantID: 1},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.ResumeShortResponse{
				{ID: 1, ApplicantID: 1},
				{ID: 2, ApplicantID: 1},
			},
		},
		{
			name:   "Success - employer role",
			cookie: &http.Cookie{Name: "session_id", Value: "employer-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "employer-session").
					Return(2, "employer", nil)
				resume.EXPECT().GetAll(gomock.Any()).
					Return([]dto.ResumeShortResponse{
						{ID: 1, ApplicantID: 1},
						{ID: 3, ApplicantID: 3},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.ResumeShortResponse{
				{ID: 1, ApplicantID: 1},
				{ID: 3, ApplicantID: 3},
			},
		},
		{
			name:           "Unauthorized - no cookie",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
		},
		{
			name:   "Unauthorized - invalid session",
			cookie: &http.Cookie{Name: "session_id", Value: "invalid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid-session").
					Return(0, "", entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.ToAPIError(entity.ErrInternal),
		},
		{
			name:   "Employer - internal error",
			cookie: &http.Cookie{Name: "session_id", Value: "employer-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "employer-session").
					Return(2, "employer", nil)
				resume.EXPECT().GetAll(gomock.Any()).
					Return(nil, entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.ToAPIError(entity.ErrInternal),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			tc.setupMocks(authMock, resumeMock)

			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

			req := httptest.NewRequest(http.MethodGet, "/resume/all", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			router := http.NewServeMux()
			handler.Configure(router)

			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp []dto.ResumeShortResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, resp)
			} else {
				var resp utils.APIError
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)

				expectedErr, ok := tc.expectedBody.(utils.APIError)
				if ok {
					require.Equal(t, expectedErr.Status, resp.Status)
					require.Equal(t, expectedErr.Message, resp.Message)
				}
			}
		})
	}
}
