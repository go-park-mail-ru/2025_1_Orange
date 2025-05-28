package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/transport/ws"
	"ResuMatch/internal/usecase/mock"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestVacancyHandler_CreateVacancy(t *testing.T) {
	t.Parallel()

	validVacancyRequest := func() *dto.VacancyCreate {
		return &dto.VacancyCreate{
			Title:                "Go Developer",
			Specialization:       "Backend",
			WorkFormat:           "Remote",
			Employment:           "Full-time",
			Schedule:             "Flexible",
			City:                 "Moscow",
			Description:          "Job description",
			Experience:           "3 years",
			Tasks:                "Develop backend services",
			Requirements:         "Go, PostgreSQL",
			OptionalRequirements: "Docker, Kubernetes",
			Skills:               []string{"Go", "PostgreSQL"},
			SalaryFrom:           100000,
			SalaryTo:             200000,
		}
	}

	tests := []struct {
		name           string
		setupMock      func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		cookie         *http.Cookie
		requestBody    interface{}
		expectedStatus int
		expectedError  error
	}{
		{
			name: "Success",
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(10, "employer", nil)
				vacancy.EXPECT().CreateVacancy(gomock.Any(), 10, gomock.Any()).Return(&dto.VacancyResponse{ID: 1}, nil)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody:    validVacancyRequest(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "No cookie - unauthorized",
			setupMock:      func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			cookie:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  entity.ErrUnauthorized,
		},
		{
			name: "Invalid session - unauthorized",
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid").Return(0, "", entity.ErrUnauthorized)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "invalid"},
			requestBody:    validVacancyRequest(),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  entity.ErrInternal,
		},
		{
			name: "Not employer - forbidden",
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(10, "applicant", nil)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody:    validVacancyRequest(),
			expectedStatus: http.StatusForbidden,
			expectedError:  entity.ErrForbidden,
		},
		{
			name: "Invalid JSON - bad request",
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil) // или другой релевантный пользователь
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody:    "{invalid json}", // строка, а не структура
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},

		{
			name: "Create vacancy error - internal",
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(10, "employer", nil)
				vacancy.EXPECT().CreateVacancy(gomock.Any(), 10, gomock.Any()).Return(nil, entity.ErrInternal)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody:    validVacancyRequest(),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  entity.ErrInternal,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockVacancy := mock.NewMockVacancy(ctrl)
			tt.setupMock(mockAuth, mockVacancy)

			handler := &VacancyHandler{
				auth:    mockAuth,
				vacancy: mockVacancy,
			}

			var body []byte
			switch v := tt.requestBody.(type) {
			case string:
				body = []byte(v)
			default:
				var err error
				body, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/vacancy/vacancies", bytes.NewReader(body))
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			w := httptest.NewRecorder()

			handler.CreateVacancy(w, req)

			resp := w.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedError != nil {
				var apiError utils.APIError
				err := json.NewDecoder(resp.Body).Decode(&apiError)
				require.NoError(t, err)
				require.Equal(t, tt.expectedError.Error(), apiError.Message)
			} else if resp.StatusCode == http.StatusCreated {
				var vacancy dto.VacancyResponse
				err := json.NewDecoder(resp.Body).Decode(&vacancy)
				require.NoError(t, err)
				require.NotEmpty(t, vacancy.ID)
			}
		})
	}
}

func TestVacancyHandler_GetVacancy(t *testing.T) {
	t.Parallel()

	// Вспомогательная функция для создания валидного ответа вакансии
	validVacancyResponse := func() *dto.VacancyResponse {
		return &dto.VacancyResponse{
			ID: 1,
			// Здесь можно добавить другие поля, если нужно
		}
	}

	tests := []struct {
		name           string
		vacancyID      string
		setupMock      func(auth *mock.MockAuth, vac *mock.MockVacancy)
		expectedStatus int
		expectedError  error
		expectedBody   *dto.VacancyResponse
	}{
		{
			name:      "Успешное получение вакансии",
			vacancyID: "1",
			setupMock: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)
				vac.EXPECT().
					GetVacancy(gomock.Any(), 1, 1, "applicant").
					Return(validVacancyResponse(), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   validVacancyResponse(),
		},
		{
			name:      "Неверный ID вакансии",
			vacancyID: "invalid",
			setupMock: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(0, "", nil) // или можно ошибку, если хотите
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:      "Ошибка при получении вакансии",
			vacancyID: "1",
			setupMock: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)
				vac.EXPECT().
					GetVacancy(gomock.Any(), 1, 1, "applicant").
					Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  entity.ErrInternal,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockVacancy := mock.NewMockVacancy(ctrl)

			tt.setupMock(mockAuth, mockVacancy)

			handler := VacancyHandler{
				auth:    mockAuth,
				vacancy: mockVacancy,
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/vacancy/vacancy/%s", tt.vacancyID), nil)
			req.SetPathValue("id", tt.vacancyID)
			req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})

			w := httptest.NewRecorder()

			handler.GetVacancy(w, req)

			resp := w.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
			require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			if tt.expectedError != nil {
				var apiError utils.APIError
				err := json.NewDecoder(resp.Body).Decode(&apiError)
				require.NoError(t, err)
				require.Equal(t, tt.expectedError.Error(), apiError.Message)
				require.Equal(t, tt.expectedStatus, apiError.Status)
			} else if resp.StatusCode == http.StatusOK {
				var vacancy dto.VacancyResponse
				err := json.NewDecoder(resp.Body).Decode(&vacancy)
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody, &vacancy)
			}
		})
	}
}

func TestVacancyHandler_UpdateVacancy(t *testing.T) {
	t.Parallel()

	validUpdate := &dto.VacancyUpdate{
		Title:                "Senior Backend Developer",
		Description:          "Work on scalable systems",
		WorkFormat:           "Remote",
		Employment:           "Full-time",
		Schedule:             "Flexible",
		City:                 "Berlin",
		Experience:           "3+ years",
		Specialization:       "Backend",
		Requirements:         "Go, SQL",
		OptionalRequirements: "Docker, Kubernetes",
		Tasks:                "Build backend services",
		Skills:               []string{"Go", "SQL"},
	}

	validVacancy := &dto.VacancyResponse{
		ID:                   1,
		EmployerID:           42,
		Title:                "Senior Backend Developer",
		Description:          "Work on scalable systems",
		WorkFormat:           "Remote",
		Employment:           "Full-time",
		Schedule:             "Flexible",
		City:                 "Berlin",
		Experience:           "3+ years",
		Specialization:       "Backend",
		Requirements:         "Go, SQL",
		OptionalRequirements: "Docker, Kubernetes",
		Tasks:                "Build backend services",
		Skills:               []string{"Go", "SQL"},
	}

	tests := []struct {
		name           string
		vacancyID      string
		cookie         *http.Cookie
		body           interface{}
		setupMock      func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedError  error
	}{
		{
			name:      "Success",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "abc123"},
			body:      validUpdate,
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "abc123").Return(42, "employer", nil)
				vacancy.EXPECT().UpdateVacancy(gomock.Any(), 1, 42, gomock.Any()).Return(validVacancy, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing cookie",
			vacancyID:      "1",
			cookie:         nil,
			body:           validUpdate,
			setupMock:      func(_ *mock.MockAuth, _ *mock.MockVacancy) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  entity.ErrUnauthorized,
		},
		{
			name:      "Invalid session",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "bad"},
			body:      validUpdate,
			setupMock: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "bad").Return(0, "", entity.ErrUnauthorized)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  entity.ErrInternal,
		},
		{
			name:      "Forbidden user type",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "abc123"},
			body:      validUpdate,
			setupMock: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "abc123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  entity.ErrForbidden,
		},
		{
			name:      "Invalid vacancy ID",
			vacancyID: "not_a_number",
			cookie:    &http.Cookie{Name: "session_id", Value: "abc123"},
			body:      validUpdate,
			setupMock: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "abc123").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:      "Invalid JSON",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "abc123"},
			body:      "{bad json",
			setupMock: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "abc123").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:      "Vacancy not found",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "abc123"},
			body:      validUpdate,
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "abc123").Return(42, "employer", nil)
				vacancy.EXPECT().UpdateVacancy(gomock.Any(), 1, 42, gomock.Any()).Return(nil, entity.NewError(entity.ErrNotFound, fmt.Errorf("vacancy not found")))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Internal error",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "abc123"},
			body:      validUpdate,
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "abc123").Return(42, "employer", nil)
				vacancy.EXPECT().UpdateVacancy(gomock.Any(), 1, 42, gomock.Any()).Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("db error")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockVacancy := mock.NewMockVacancy(ctrl)

			if tt.setupMock != nil {
				tt.setupMock(mockAuth, mockVacancy)
			}

			handler := &VacancyHandler{
				auth:    mockAuth,
				vacancy: mockVacancy,
			}

			var body []byte
			var err error

			switch v := tt.body.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(v)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/vacancy/%s", tt.vacancyID), bytes.NewReader(body))
			req.SetPathValue("id", tt.vacancyID)

			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			rec := httptest.NewRecorder()
			handler.UpdateVacancy(rec, req)

			resp := rec.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus != http.StatusOK {
				var apiErr utils.APIError
				_ = json.NewDecoder(resp.Body).Decode(&apiErr)
				require.Equal(t, tt.expectedStatus, apiErr.Status)
			} else {
				var result dto.VacancyResponse
				err := json.NewDecoder(resp.Body).Decode(&result)
				require.NoError(t, err)
				require.Equal(t, validVacancy.ID, result.ID)
			}
		})
	}
}

func TestVacancyHandler_DeleteVacancy(t *testing.T) {
	t.Parallel()

	validDeleteVacancyResponse := func() *dto.DeleteVacancy {
		return &dto.DeleteVacancy{
			Success: true,
		}
	}

	tests := []struct {
		name           string
		vacancyID      string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedError  error
		expectedBody   *dto.DeleteVacancy
	}{
		{
			name:      "Success",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)
				vacancy.EXPECT().
					DeleteVacancy(gomock.Any(), 1, 1).
					Return(validDeleteVacancyResponse(), nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   validDeleteVacancyResponse(),
		},
		{
			name:           "No cookie - unauthorized",
			vacancyID:      "1",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  entity.ErrUnauthorized,
		},
		{
			name:      "Invalid session - unauthorized",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "invalid"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "invalid").
					Return(0, "", entity.NewError(entity.ErrUnauthorized, fmt.Errorf("invalid session")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  fmt.Errorf("invalid session"),
		},
		{
			name:      "Forbidden - user not employer",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  entity.ErrForbidden,
		},
		{
			name:      "Invalid vacancy ID - bad request",
			vacancyID: "invalid",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:      "Vacancy not found",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)
				vacancy.EXPECT().
					DeleteVacancy(gomock.Any(), 1, 1).
					Return(nil, entity.NewError(entity.ErrNotFound, fmt.Errorf("vacancy not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  fmt.Errorf("vacancy not found"),
		},
		{
			name:      "Not owner - forbidden",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)
				vacancy.EXPECT().
					DeleteVacancy(gomock.Any(), 1, 1).
					Return(nil, entity.NewError(entity.ErrForbidden, fmt.Errorf("not authorized to delete vacancy")))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  fmt.Errorf("not authorized to delete vacancy"),
		},
		{
			name:      "Internal server error",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)
				vacancy.EXPECT().
					DeleteVacancy(gomock.Any(), 1, 1).
					Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("database error")))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockVacancy := mock.NewMockVacancy(ctrl)

			tt.setupMocks(mockAuth, mockVacancy)

			handler := &VacancyHandler{
				auth:    mockAuth,
				vacancy: mockVacancy,
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/vacancy/%s", tt.vacancyID), nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			req.SetPathValue("id", tt.vacancyID)

			w := httptest.NewRecorder()

			handler.DeleteVacancy(w, req)

			resp := w.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
			require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			if tt.expectedError != nil {
				var apiError utils.APIError
				err := json.NewDecoder(resp.Body).Decode(&apiError)
				require.NoError(t, err)
				require.Equal(t, tt.expectedError.Error(), apiError.Message)
				require.Equal(t, tt.expectedStatus, apiError.Status)
			} else if resp.StatusCode == http.StatusOK {
				var response dto.DeleteVacancy
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, tt.expectedBody, &response)
			}
		})
	}
}

func TestVacancyHandler_GetAllVacancies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(*mock.MockAuth, *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Успешное получение вакансий",
			queryParams: map[string]string{
				"limit":  "10",
				"offset": "0",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)

				vac.EXPECT().
					GetAll(gomock.Any(), 1, "applicant", 10, 0).
					Return([]dto.VacancyShortResponse{{ID: 1}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []dto.VacancyShortResponse{{ID: 1}},
		},
		{
			name: "Неверные параметры пагинации",
			queryParams: map[string]string{
				"limit":  "invalid",
				"offset": "invalid",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   entity.ErrBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockVacancy := mock.NewMockVacancy(ctrl)

			tc.mockSetup(mockAuth, mockVacancy)

			handler := VacancyHandler{
				auth:    mockAuth,
				vacancy: mockVacancy,
			}

			req := httptest.NewRequest("GET", "/vacancies", nil)
			q := req.URL.Query()
			for k, v := range tc.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
			w := httptest.NewRecorder()

			handler.GetAllVacancies(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response []dto.VacancyShortResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, response)
			}
		})
	}
}

func TestVacancyHandler_ApplyToVacancy(t *testing.T) {
	t.Parallel()

	type applicationBody struct {
		Message string `json:"message"`
	}

	tests := []struct {
		name           string
		vacancyID      string
		resumeID       string
		cookie         *http.Cookie
		body           interface{}
		setupMock      func(auth *mock.MockAuth, vacancy *mock.MockVacancy, notif *mock.MockNotification)
		expectedStatus int
	}{
		// {
		// 	name:      "Success",
		// 	vacancyID: "1",
		// 	resumeID:  "2",
		// 	cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
		// 	body:      applicationBody{Message: "I am interested!"},
		// 	setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy, notif *mock.MockNotification) {
		// 		auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
		// 		vacancy.EXPECT().ApplyToVacancy(gomock.Any(), 1, 1, 2).Return(entity.Notification{
		// 			ReceiverID: 2,
		// 			SenderID:   1,
		// 			Type:       "application",
		// 		}, nil)
		// 		notif.EXPECT().CreateNotification(gomock.Any(), gomock.Any()).Return(&entity.NotificationPreview{}, nil)
		// 	},
		// 	expectedStatus: http.StatusCreated,
		// },
		{
			name:           "No cookie - unauthorized",
			vacancyID:      "1",
			resumeID:       "2",
			cookie:         nil,
			body:           nil,
			setupMock:      func(_ *mock.MockAuth, _ *mock.MockVacancy, _ *mock.MockNotification) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "Invalid vacancy ID - bad request",
			vacancyID: "abc",
			resumeID:  "2",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			body:      nil,
			setupMock: func(auth *mock.MockAuth, _ *mock.MockVacancy, _ *mock.MockNotification) {
				//auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Invalid resume ID - bad request",
			vacancyID: "1",
			resumeID:  "xyz",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			body:      nil,
			setupMock: func(auth *mock.MockAuth, _ *mock.MockVacancy, _ *mock.MockNotification) {
				//auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Not applicant - forbidden",
			vacancyID: "1",
			resumeID:  "2",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			body:      nil,
			setupMock: func(auth *mock.MockAuth, _ *mock.MockVacancy, _ *mock.MockNotification) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:      "ApplyToVacancy returns error",
			vacancyID: "1",
			resumeID:  "2",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			body:      nil,
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy, _ *mock.MockNotification) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				vacancy.EXPECT().
					ApplyToVacancy(gomock.Any(), 1, 1, 2).
					Return(entity.Notification{}, entity.NewError(entity.ErrNotFound, errors.New("vacancy not found")))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "Notification creation error",
			vacancyID: "1",
			resumeID:  "2",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			body:      nil,
			setupMock: func(auth *mock.MockAuth, vacancy *mock.MockVacancy, notif *mock.MockNotification) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				vacancy.EXPECT().
					ApplyToVacancy(gomock.Any(), 1, 1, 2).
					Return(entity.Notification{}, nil)
				notif.EXPECT().
					CreateNotification(gomock.Any(), gomock.Any()).
					Return(nil, entity.NewError(entity.ErrInternal, errors.New("notification error")))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			vacancyMock := mock.NewMockVacancy(ctrl)
			chatMock := mock.NewMockChat(ctrl)
			wsHub := ws.NewHub(chatMock)
			notificationMock := mock.NewMockNotification(ctrl)

			tt.setupMock(authMock, vacancyMock, notificationMock)

			handler := NewVacancyHandler(authMock, vacancyMock, config.CSRFConfig{}, wsHub, notificationMock)

			var body []byte
			if tt.body != nil {
				body, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost,
				fmt.Sprintf("/vacancy/%s/response/%s", tt.vacancyID, tt.resumeID),
				bytes.NewReader(body))
			req.SetPathValue("id", tt.vacancyID)
			req.SetPathValue("resume_id", tt.resumeID)

			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			w := httptest.NewRecorder()
			handler.ApplyToVacancy(w, req)

			resp := w.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestVacancyHandler_SearchVacancies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func(*mock.MockAuth, *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Успешный поиск вакансий",
			queryParams: map[string]string{
				"query":  "developer",
				"limit":  "10",
				"offset": "0",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)

				vac.EXPECT().
					SearchVacancies(gomock.Any(), 1, "applicant", "developer", 10, 0).
					Return([]dto.VacancyShortResponse{{ID: 1}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   []dto.VacancyShortResponse{{ID: 1}},
		},
		{
			name: "Отсутствует параметр query",
			queryParams: map[string]string{
				"limit":  "10",
				"offset": "0",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(0, "", errors.New("no session or invalid"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockVacancy := mock.NewMockVacancy(ctrl)

			tc.mockSetup(mockAuth, mockVacancy)

			handler := VacancyHandler{
				auth:    mockAuth,
				vacancy: mockVacancy,
			}

			req := httptest.NewRequest("GET", "/search", nil)
			q := req.URL.Query()
			for k, v := range tc.queryParams {
				q.Add(k, v)
			}
			req.URL.RawQuery = q.Encode()
			req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
			w := httptest.NewRecorder()

			handler.SearchVacancies(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response []dto.VacancyShortResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, response)
			}
		})
	}
}

func TestVacancyHandler_SearchVacanciesBySpecializations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		requestBody    interface{}
		queryParams    string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success with auth",
			requestBody: dto.SearchBySpecializationsRequest{
				Specializations: []string{"Backend", "Frontend"},
			},
			queryParams: "?limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				vacancy.EXPECT().SearchVacanciesBySpecializations(
					gomock.Any(), 1, "applicant", []string{"Backend", "Frontend"}, 10, 0).
					Return([]dto.VacancyShortResponse{
						{ID: 1, Title: "Backend Developer"},
						{ID: 2, Title: "Frontend Developer"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.VacancyShortResponse{
				{ID: 1, Title: "Backend Developer"},
				{ID: 2, Title: "Frontend Developer"},
			},
		},
		{
			name: "Success without auth",
			requestBody: dto.SearchBySpecializationsRequest{
				Specializations: []string{"DevOps"},
			},
			queryParams: "?limit=5&offset=0",
			cookie:      nil,
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				vacancy.EXPECT().SearchVacanciesBySpecializations(
					gomock.Any(), 0, "", []string{"DevOps"}, 5, 0).
					Return([]dto.VacancyShortResponse{
						{ID: 3, Title: "DevOps Engineer"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.VacancyShortResponse{
				{ID: 3, Title: "DevOps Engineer"},
			},
		},
		{
			name: "Bad request - empty specializations",
			requestBody: dto.SearchBySpecializationsRequest{
				Specializations: []string{},
			},
			queryParams:    "",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: "список специализаций не может быть пустым"},
		},
		{
			name:           "Bad request - invalid body",
			requestBody:    "invalid body",
			queryParams:    "",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
		},
		{
			name: "Internal server error",
			requestBody: dto.SearchBySpecializationsRequest{
				Specializations: []string{"Backend"},
			},
			queryParams: "?limit=10&offset=0",
			cookie:      nil,
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				vacancy.EXPECT().SearchVacanciesBySpecializations(
					gomock.Any(), 0, "", []string{"Backend"}, 10, 0).
					Return(nil, entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.APIError{Status: http.StatusInternalServerError, Message: entity.ErrInternal.Error()},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			vacancyMock := mock.NewMockVacancy(ctrl)
			chatMock := mock.NewMockChat(ctrl)
			wsHub := ws.NewHub(chatMock)
			notificationMock := mock.NewMockNotification(ctrl)

			tc.setupMocks(authMock, vacancyMock)

			handler := NewVacancyHandler(authMock, vacancyMock, config.CSRFConfig{}, wsHub, notificationMock)

			var reqBody []byte
			switch body := tc.requestBody.(type) {
			case string:
				reqBody = []byte(body)
			default:
				var err error
				reqBody, err = json.Marshal(body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/vacancy/search/specializations"+tc.queryParams, bytes.NewReader(reqBody))
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			rec := httptest.NewRecorder()

			handler.SearchVacanciesBySpecializations(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp []dto.VacancyShortResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, resp)
			} else {
				var resp utils.APIError
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, resp)
			}
		})
	}
}

func TestVacancyHandler_GetActiveVacanciesByEmployer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		employerID     string
		limit          string
		offset         string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "Success - with session",
			employerID: "5",
			cookie:     &http.Cookie{Name: "session_id", Value: "session123"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(42, "employer", nil)
				vacancy.EXPECT().GetActiveVacanciesByEmployerID(gomock.Any(), 5, 42, "employer", 10, 0).
					Return([]dto.VacancyShortResponse{{ID: 1, Title: "Backend Go"}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"city":"", "created_at":"", "employer":null, "employment":"", "id":1, "liked":false, "responded":false, "salary_from":0, "salary_to":0, "specialization":"", "taxes_included":false, "title":"Backend Go", "updated_at":"", "work_format":"", "working_hours":0}]`,
		},
		{
			name:       "Success - without session (guest)",
			employerID: "5",
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				vacancy.EXPECT().GetActiveVacanciesByEmployerID(gomock.Any(), 5, 0, "", 10, 0).
					Return([]dto.VacancyShortResponse{{ID: 2, Title: "Frontend Vue"}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"city":"", "created_at":"", "employer":null, "employment":"", "id":2, "liked":false, "responded":false, "salary_from":0, "salary_to":0, "specialization":"", "taxes_included":false, "title":"Frontend Vue", "updated_at":"", "work_format":"", "working_hours":0}]`,
		},
		{
			name:           "Invalid employer ID",
			employerID:     "bad_id",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid limit",
			employerID: "1",
			limit:      "invalid",
			cookie:     &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid offset",
			employerID: "1",
			offset:     "invalid",
			cookie:     &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "GetActiveVacanciesByEmployerID returns error",
			employerID: "1",
			cookie:     &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "employer", nil)
				vacancy.EXPECT().GetActiveVacanciesByEmployerID(gomock.Any(), 1, 1, "employer", 10, 0).
					Return(nil, errors.New("db failure"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			vacancyMock := mock.NewMockVacancy(ctrl)
			chatMock := mock.NewMockChat(ctrl)
			wsHub := ws.NewHub(chatMock)
			notificationMock := mock.NewMockNotification(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(authMock, vacancyMock)
			}

			handler := NewVacancyHandler(authMock, vacancyMock, config.CSRFConfig{}, wsHub, notificationMock)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/vacancy/employer/%s/active", tt.employerID), nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			req.SetPathValue("id", tt.employerID)

			q := req.URL.Query()
			if tt.limit != "" {
				q.Set("limit", tt.limit)
			}
			if tt.offset != "" {
				q.Set("offset", tt.offset)
			}
			req.URL.RawQuery = q.Encode()

			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			rec := httptest.NewRecorder()
			handler.GetActiveVacanciesByEmployer(rec, req)

			resp := rec.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()
			body, _ := io.ReadAll(resp.Body)

			require.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedBody != "" {
				require.JSONEq(t, tt.expectedBody, string(body))
			}
		})
	}
}

func TestVacancyHandler_GetResponsesOnVacancy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		vacancyID      string
		cookie         *http.Cookie
		userType       string
		mockSetup      func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		limit          string
		offset         string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Success",
			vacancyID: "123",
			cookie:    &http.Cookie{Name: "session_id", Value: "session123"},
			userType:  "employer",
			mockSetup: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "employer", nil)
				vacancy.EXPECT().GetRespondedResumeOnVacancy(gomock.Any(), 123, 10, 0).Return([]dto.ResumeApplicantShortResponse{
					{ID: 1, Specialization: "Developer"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: `[{
			"applicant": null,
			"created_at": "",
			"id": 1,
			"profession": "",
			"skills": null,
			"specialization": "Developer",
			"updated_at": "",
			"work_experiences": {
				"employer_name": "",
				"id": 0,
				"position": "",
				"start_date": "",
				"until_now": false
			}
			}]`,
		},
		{
			name:           "Unauthorized - no cookie",
			vacancyID:      "123",
			cookie:         nil,
			mockSetup:      func(_ *mock.MockAuth, _ *mock.MockVacancy) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:      "Invalid vacancy ID",
			vacancyID: "abc",
			cookie:    &http.Cookie{Name: "session_id", Value: "s"},
			mockSetup: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				// НЕ ожидаем вызова GetUserIDBySession, потому что он не должен быть вызван при невалидном ID
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Invalid limit",
			vacancyID: "1",
			limit:     "bad",
			cookie:    &http.Cookie{Name: "session_id", Value: "s"},
			mockSetup: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Invalid offset",
			vacancyID: "1",
			offset:    "bad",
			cookie:    &http.Cookie{Name: "session_id", Value: "s"},
			mockSetup: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Forbidden - user not employer",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "s"},
			mockSetup: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:      "Auth error",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "s"},
			mockSetup: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(0, "", errors.New("session not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:      "Vacancy usecase returns error",
			vacancyID: "1",
			cookie:    &http.Cookie{Name: "session_id", Value: "s"},
			userType:  "employer",
			mockSetup: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "employer", nil)
				vacancy.EXPECT().GetRespondedResumeOnVacancy(gomock.Any(), 1, 10, 0).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockVacancy := mock.NewMockVacancy(ctrl)

			if tt.mockSetup != nil {
				tt.mockSetup(mockAuth, mockVacancy)
			}

			handler := &VacancyHandler{
				auth:    mockAuth,
				vacancy: mockVacancy,
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/vacancy/%s/responses", tt.vacancyID), nil)
			req.SetPathValue("id", tt.vacancyID)

			q := req.URL.Query()
			if tt.limit != "" {
				q.Set("limit", tt.limit)
			}
			if tt.offset != "" {
				q.Set("offset", tt.offset)
			}
			req.URL.RawQuery = q.Encode()

			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			w := httptest.NewRecorder()
			handler.GetResponsesOnVacancy(w, req)

			resp := w.Result()
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()

			body, _ := io.ReadAll(resp.Body)

			require.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				require.JSONEq(t, tt.expectedBody, string(body))
			}
		})
	}
}

func TestVacancyHandler_GetVacanciesByApplicant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pathID         string
		cookie         *http.Cookie
		limit          string
		offset         string
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			pathID: "1",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "applicant", nil)
				vacancy.EXPECT().GetVacanciesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]dto.VacancyShortResponse{{ID: 7, Title: "Go Dev"}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"city":"", "created_at":"", "employer": null, "employment":"", "id":7, "liked":false, "responded":false, "salary_from":0, "salary_to":0, "specialization":"", "taxes_included":false, "title":"Go Dev", "updated_at":"", "work_format":"", "working_hours":0}]`,
		},
		{
			name:           "No cookie",
			pathID:         "1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "GetUserIDBySession error",
			pathID: "1",
			cookie: &http.Cookie{Name: "session_id", Value: "bad"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "bad").Return(0, "", errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Forbidden - wrong role",
			pathID: "1",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "Invalid applicantID in path",
			pathID: "abc",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Forbidden - applicantID mismatch",
			pathID: "999",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "Invalid limit param",
			pathID: "1",
			limit:  "not-a-number",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid offset param",
			pathID: "1",
			offset: "oops",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Vacancy service error",
			pathID: "1",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(1, "applicant", nil)
				vacancy.EXPECT().GetVacanciesByApplicantID(gomock.Any(), 1, 10, 0).Return(nil, errors.New("fail"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockVacancy := mock.NewMockVacancy(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(mockAuth, mockVacancy)
			}

			handler := &VacancyHandler{
				auth:    mockAuth,
				vacancy: mockVacancy,
			}

			req := httptest.NewRequest(http.MethodGet, "/applicant/"+tt.pathID+"/vacancies", nil)
			req.SetPathValue("id", tt.pathID)
			if tt.limit != "" || tt.offset != "" {
				q := req.URL.Query()
				if tt.limit != "" {
					q.Set("limit", tt.limit)
				}
				if tt.offset != "" {
					q.Set("offset", tt.offset)
				}
				req.URL.RawQuery = q.Encode()
			}
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			rec := httptest.NewRecorder()
			handler.GetVacanciesByApplicant(rec, req)

			res := rec.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()
			body, _ := io.ReadAll(res.Body)

			require.Equal(t, tt.expectedStatus, res.StatusCode)
			if tt.expectedBody != "" {
				require.JSONEq(t, tt.expectedBody, string(body))
			}
		})
	}
}

func TestVacancyHandler_GetLikedVacancies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pathID         string
		cookie         *http.Cookie
		limit          string
		offset         string
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			pathID: "42",
			cookie: &http.Cookie{Name: "session_id", Value: "session123"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").
					Return(42, "applicant", nil)
				vacancy.EXPECT().GetLikedVacancies(gomock.Any(), 42, 10, 0).
					Return([]dto.VacancyShortResponse{{ID: 1, Title: "Backend Go"}}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"city":"", "created_at":"", "employer":null, "employment":"", "id":1, "liked":false, "responded":false, "salary_from":0, "salary_to":0, "specialization":"", "taxes_included":false, "title":"Backend Go", "updated_at":"", "work_format":"", "working_hours":0}]`,
		},
		{
			name:           "Unauthorized - no cookie",
			pathID:         "42",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "GetUserIDBySession error",
			pathID: "42",
			cookie: &http.Cookie{Name: "session_id", Value: "bad"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "bad").
					Return(0, "", errors.New("session error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "Forbidden - wrong role",
			pathID: "42",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").
					Return(42, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "Invalid applicant ID",
			pathID: "not-an-int",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(42, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Forbidden - applicantID != currentUserID",
			pathID: "999",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(42, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "Invalid limit",
			pathID: "42",
			limit:  "bad",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(42, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Invalid offset",
			pathID: "42",
			offset: "bad",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, _ *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(42, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GetLikedVacancies returns error",
			pathID: "42",
			cookie: &http.Cookie{Name: "session_id", Value: "s"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "s").Return(42, "applicant", nil)
				vacancy.EXPECT().GetLikedVacancies(gomock.Any(), 42, 10, 0).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			vacancyMock := mock.NewMockVacancy(ctrl)
			chatMock := mock.NewMockChat(ctrl)
			wsHub := ws.NewHub(chatMock)
			notificationMock := mock.NewMockNotification(ctrl)

			if tt.setupMocks != nil {
				tt.setupMocks(authMock, vacancyMock)
			}

			handler := NewVacancyHandler(authMock, vacancyMock, config.CSRFConfig{}, wsHub, notificationMock)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/vacancy/applicant/%s/liked", tt.pathID), nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			req.SetPathValue("id", tt.pathID)

			q := req.URL.Query()
			if tt.limit != "" {
				q.Set("limit", tt.limit)
			}
			if tt.offset != "" {
				q.Set("offset", tt.offset)
			}
			req.URL.RawQuery = q.Encode()

			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			rec := httptest.NewRecorder()
			handler.GetLikedVacancies(rec, req)

			res := rec.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					t.Errorf("Failed to close response body: %v", err)
				}
			}()
			body, _ := io.ReadAll(res.Body)

			require.Equal(t, tt.expectedStatus, res.StatusCode)
			if tt.expectedBody != "" {
				require.JSONEq(t, tt.expectedBody, string(body))
			}
		})
	}
}

func TestVacancyHandler_LikeVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		vacancyID      string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedError  error
	}{
		{
			name:      "Success",
			vacancyID: "123",
			cookie:    &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				vacancy.EXPECT().LikeVacancy(gomock.Any(), 123, 1).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:      "Forbidden - not applicant",
			vacancyID: "123",
			cookie:    &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:      "Bad request - invalid vacancy ID",
			vacancyID: "invalid",
			cookie:    &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(0, "", nil).AnyTimes() // чтобы игнорировать вызовы
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  entity.ErrForbidden,
		},
		{
			name:      "Internal server error",
			vacancyID: "123",
			cookie:    &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				vacancy.EXPECT().LikeVacancy(gomock.Any(), 123, 1).
					Return(entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  entity.ErrInternal,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authMock := mock.NewMockAuth(ctrl)
			vacancyMock := mock.NewMockVacancy(ctrl)
			chatMock := mock.NewMockChat(ctrl)
			wsHub := ws.NewHub(chatMock)
			notificationMock := mock.NewMockNotification(ctrl)
			tc.setupMocks(authMock, vacancyMock)

			handler := NewVacancyHandler(authMock, vacancyMock, config.CSRFConfig{}, wsHub, notificationMock)

			req := httptest.NewRequest(http.MethodPost, "/vacancy/"+tc.vacancyID+"/like", nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}
			req.SetPathValue("id", tc.vacancyID)

			rec := httptest.NewRecorder()

			handler.LikeVacancy(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedError != nil {
				var apiErr utils.APIError
				err := json.NewDecoder(rec.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tc.expectedError.Error(), apiErr.Message)
				require.Equal(t, tc.expectedStatus, apiErr.Status)
			}
		})
	}
}
