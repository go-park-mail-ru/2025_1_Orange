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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestVacancyHandler_CreateVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*mock.MockAuth, *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Успешное создание вакансии",
			requestBody: dto.VacancyCreate{
				Title: "Backend Developer",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)

				vac.EXPECT().
					CreateVacancy(gomock.Any(), 1, gomock.Any()).
					Return(&dto.VacancyResponse{ID: 1, Title: "Backend Developer"}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   &dto.VacancyResponse{ID: 1, Title: "Backend Developer"},
		},
		{
			name:           "Неверный формат запроса",
			requestBody:    "invalid",
			mockSetup:      func(auth *mock.MockAuth, vac *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   entity.ErrBadRequest,
		},
		{
			name: "Пользователь не работодатель",
			requestBody: dto.VacancyCreate{
				Title: "Backend Developer",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   entity.ErrForbidden,
		},
		{
			name: "Ошибка при создании вакансии",
			requestBody: dto.VacancyCreate{
				Title: "Backend Developer",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)

				vac.EXPECT().
					CreateVacancy(gomock.Any(), 1, gomock.Any()).
					Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   entity.ErrInternal,
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

			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("POST", "/vacancies", bytes.NewReader(body))
			req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
			w := httptest.NewRecorder()

			handler.CreateVacancy(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusCreated {
				var response dto.VacancyResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, &response)
			}
		})
	}
}

func TestVacancyHandler_GetVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		vacancyID      string
		mockSetup      func(*mock.MockAuth, *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:      "Успешное получение вакансии",
			vacancyID: "1",
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)

				vac.EXPECT().
					GetVacancy(gomock.Any(), 1, 1, "applicant").
					Return(&dto.VacancyResponse{ID: 1}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &dto.VacancyResponse{ID: 1},
		},
		{
			name:           "Неверный ID вакансии",
			vacancyID:      "invalid",
			mockSetup:      func(auth *mock.MockAuth, vac *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   entity.ErrBadRequest,
		},
		{
			name:      "Ошибка при получении вакансии",
			vacancyID: "1",
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)

				vac.EXPECT().
					GetVacancy(gomock.Any(), 1, 1, "applicant").
					Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   entity.ErrInternal,
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

			req := httptest.NewRequest("GET", "/vacancy/"+tc.vacancyID, nil)
			req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
			w := httptest.NewRecorder()

			handler.GetVacancy(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response dto.VacancyResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, &response)
			}
		})
	}
}

func TestVacancyHandler_UpdateVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		vacancyID      string
		requestBody    interface{}
		mockSetup      func(*mock.MockAuth, *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:      "Успешное обновление вакансии",
			vacancyID: "1",
			requestBody: dto.VacancyUpdate{
				Title: "Updated Title",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)

				vac.EXPECT().
					UpdateVacancy(gomock.Any(), 1, 1, gomock.Any()).
					Return(&dto.VacancyResponse{ID: 1, Title: "Updated Title"}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &dto.VacancyResponse{ID: 1, Title: "Updated Title"},
		},
		{
			name:           "Неверный формат запроса",
			vacancyID:      "1",
			requestBody:    "invalid",
			mockSetup:      func(auth *mock.MockAuth, vac *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   entity.ErrBadRequest,
		},
		{
			name:      "Пользователь не работодатель",
			vacancyID: "1",
			requestBody: dto.VacancyUpdate{
				Title: "Updated Title",
			},
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   entity.ErrForbidden,
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

			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("PUT", "/vacancy/"+tc.vacancyID, bytes.NewReader(body))
			req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
			w := httptest.NewRecorder()

			handler.UpdateVacancy(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response dto.VacancyResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, &response)
			}
		})
	}
}

func TestVacancyHandler_DeleteVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		vacancyID      string
		mockSetup      func(*mock.MockAuth, *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:      "Успешное удаление вакансии",
			vacancyID: "1",
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)

				vac.EXPECT().
					DeleteVacancy(gomock.Any(), 1, 1).
					Return(&dto.DeleteVacancy{Success: true}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   &dto.DeleteVacancy{Success: true},
		},
		{
			name:           "Неверный ID вакансии",
			vacancyID:      "invalid",
			mockSetup:      func(auth *mock.MockAuth, vac *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   entity.ErrBadRequest,
		},
		{
			name:      "Пользователь не работодатель",
			vacancyID: "1",
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   entity.ErrForbidden,
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

			req := httptest.NewRequest("DELETE", "/vacancy/"+tc.vacancyID, nil)
			req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
			w := httptest.NewRecorder()

			handler.DeleteVacancy(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedStatus == http.StatusOK {
				var response dto.DeleteVacancy
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, &response)
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
			mockSetup:      func(auth *mock.MockAuth, vac *mock.MockVacancy) {},
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

	testCases := []struct {
		name           string
		vacancyID      string
		mockSetup      func(*mock.MockAuth, *mock.MockVacancy)
		expectedStatus int
	}{
		{
			name:      "Успешный отклик на вакансию",
			vacancyID: "1",
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "applicant", nil)

				vac.EXPECT().
					ApplyToVacancy(gomock.Any(), 1, 1, 1).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Неверный ID вакансии",
			vacancyID:      "invalid",
			mockSetup:      func(auth *mock.MockAuth, vac *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Пользователь не соискатель",
			vacancyID: "1",
			mockSetup: func(auth *mock.MockAuth, vac *mock.MockVacancy) {
				auth.EXPECT().
					GetUserIDBySession(gomock.Any(), "session123").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
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

			req := httptest.NewRequest("POST", "/vacancy/"+tc.vacancyID+"/response", nil)
			req.AddCookie(&http.Cookie{Name: "session_id", Value: "session123"})
			w := httptest.NewRecorder()

			handler.ApplyToVacancy(w, req)

			require.Equal(t, tc.expectedStatus, w.Code)
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
			mockSetup:      func(auth *mock.MockAuth, vac *mock.MockVacancy) {},
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
					Return([]dto.VacancyResponse{
						{ID: 1, Title: "Backend Developer"},
						{ID: 2, Title: "Frontend Developer"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.VacancyResponse{
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
					Return([]dto.VacancyResponse{
						{ID: 3, Title: "DevOps Engineer"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.VacancyResponse{
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
				var resp []dto.VacancyResponse
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

func TestVacancyHandler_SearchVacanciesByQueryAndSpecializations(t *testing.T) {
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
			requestBody: dto.SearchByQueryAndSpecializationsRequest{
				Specializations: []string{"Backend", "Frontend"},
			},
			queryParams: "?query=developer&limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				vacancy.EXPECT().SearchVacanciesByQueryAndSpecializations(
					gomock.Any(), 1, "applicant", "developer", []string{"Backend", "Frontend"}, 10, 0).
					Return([]dto.VacancyResponse{
						{ID: 1, Title: "Senior Backend Developer"},
						{ID: 2, Title: "Junior Frontend Developer"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.VacancyResponse{
				{ID: 1, Title: "Senior Backend Developer"},
				{ID: 2, Title: "Junior Frontend Developer"},
			},
		},
		{
			name: "Bad request - empty query",
			requestBody: dto.SearchByQueryAndSpecializationsRequest{
				Specializations: []string{"DevOps"},
			},
			queryParams:    "?limit=5&offset=0",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: "параметр query обязателен"},
		},
		{
			name: "Bad request - empty specializations",
			requestBody: dto.SearchByQueryAndSpecializationsRequest{
				Specializations: []string{},
			},
			queryParams:    "?query=devops",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: "список специализаций не может быть пустым"},
		},
		{
			name: "Internal server error",
			requestBody: dto.SearchByQueryAndSpecializationsRequest{
				Specializations: []string{"Backend"},
			},
			queryParams: "?query=golang&limit=10&offset=0",
			cookie:      nil,
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				vacancy.EXPECT().SearchVacanciesByQueryAndSpecializations(
					gomock.Any(), 0, "", "golang", []string{"Backend"}, 10, 0).
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

			req := httptest.NewRequest(http.MethodPost, "/vacancy/search/query"+tc.queryParams, bytes.NewReader(reqBody))
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			rec := httptest.NewRecorder()

			handler.SearchVacanciesByQueryAndSpecializations(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp []dto.VacancyResponse
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

	testCases := []struct {
		name           string
		url            string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "Success with auth",
			url:    "/vacancy/employer/123?limit=10&offset=0",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(456, "employer", nil)
				vacancy.EXPECT().GetActiveVacanciesByEmployerID(
					gomock.Any(), 123, 456, "employer", 10, 0).
					Return([]dto.VacancyResponse{
						{ID: 1, Title: "Backend Developer"},
						{ID: 2, Title: "Frontend Developer"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.VacancyResponse{
				{ID: 1, Title: "Backend Developer"},
				{ID: 2, Title: "Frontend Developer"},
			},
		},
		{
			name:   "Success without auth",
			url:    "/vacancy/employer/123?limit=5&offset=0",
			cookie: nil,
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				vacancy.EXPECT().GetActiveVacanciesByEmployerID(
					gomock.Any(), 123, 0, "", 5, 0).
					Return([]dto.VacancyResponse{
						{ID: 3, Title: "DevOps Engineer"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.VacancyResponse{
				{ID: 3, Title: "DevOps Engineer"},
			},
		},
		{
			name:           "Bad request - invalid employer ID",
			url:            "/vacancy/employer/invalid",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
		},
		{
			name:           "Bad request - invalid limit",
			url:            "/vacancy/employer/123?limit=invalid",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
		},
		{
			name:   "Internal server error",
			url:    "/vacancy/employer/123?limit=10&offset=0",
			cookie: nil,
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				vacancy.EXPECT().GetActiveVacanciesByEmployerID(
					gomock.Any(), 123, 0, "", 10, 0).
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

			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			// Set path param
			if strings.Contains(tc.url, "/123") {
				req = mux.SetURLVars(req, map[string]string{"id": "123"})
			} else if strings.Contains(tc.url, "/invalid") {
				req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
			}

			rec := httptest.NewRecorder()

			handler.GetActiveVacanciesByEmployer(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp []dto.VacancyResponse
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

func TestVacancyHandler_GetLikedVacancies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		url            string
		queryParams    string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "Success",
			url:         "/vacancy/liked/123",
			queryParams: "?limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(123, "applicant", nil)
				vacancy.EXPECT().GetLikedVacancies(gomock.Any(), 123, 10, 0).
					Return([]dto.VacancyResponse{
						{ID: 1, Title: "Backend Developer"},
						{ID: 2, Title: "Frontend Developer"},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []dto.VacancyResponse{
				{ID: 1, Title: "Backend Developer"},
				{ID: 2, Title: "Frontend Developer"},
			},
		},
		{
			name:        "Forbidden - not applicant",
			url:         "/vacancy/liked/123",
			queryParams: "",
			cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(123, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   utils.APIError{Status: http.StatusForbidden, Message: entity.ErrForbidden.Error()},
		},
		{
			name:        "Forbidden - wrong user",
			url:         "/vacancy/liked/456",
			queryParams: "",
			cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(123, "applicant", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   utils.APIError{Status: http.StatusForbidden, Message: entity.ErrForbidden.Error()},
		},
		{
			name:           "Unauthorized - no cookie",
			url:            "/vacancy/liked/123",
			queryParams:    "",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
		},
		{
			name:        "Internal server error",
			url:         "/vacancy/liked/123",
			queryParams: "?limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(123, "applicant", nil)
				vacancy.EXPECT().GetLikedVacancies(gomock.Any(), 123, 10, 0).
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

			req := httptest.NewRequest(http.MethodGet, tc.url+tc.queryParams, nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			// Исправление: проверяем наличие id в URL перед установкой
			if strings.Contains(tc.url, "123") {
				req = mux.SetURLVars(req, map[string]string{"id": "123"})
			} else if strings.Contains(tc.url, "456") {
				req = mux.SetURLVars(req, map[string]string{"id": "456"})
			}

			rec := httptest.NewRecorder()

			handler.GetLikedVacancies(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

			if tc.expectedStatus == http.StatusOK {
				var resp []dto.VacancyResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, resp)
			} else if tc.expectedStatus != http.StatusCreated && tc.expectedStatus != http.StatusNoContent {
				var resp utils.APIError
				err := json.NewDecoder(rec.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, resp)
			}
		})
	}
}

func TestVacancyHandler_LikeVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		url            string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, vacancy *mock.MockVacancy)
		expectedStatus int
	}{
		{
			name:   "Success",
			url:    "/vacancy/like/123",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				vacancy.EXPECT().LikeVacancy(gomock.Any(), 123, 1).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Forbidden - not applicant",
			url:    "/vacancy/like/123",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Unauthorized - no cookie",
			url:            "/vacancy/like/123",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Bad request - invalid vacancy ID",
			url:    "/vacancy/like/invalid",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Internal server error",
			url:    "/vacancy/like/123",
			cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
			setupMocks: func(auth *mock.MockAuth, vacancy *mock.MockVacancy) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
				vacancy.EXPECT().LikeVacancy(gomock.Any(), 123, 1).
					Return(entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
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

			req := httptest.NewRequest(http.MethodPost, tc.url, nil)
			if tc.cookie != nil {
				req.AddCookie(tc.cookie)
			}

			// Set path param
			req = mux.SetURLVars(req, map[string]string{"id": "123"})

			rec := httptest.NewRecorder()

			handler.LikeVacancy(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}
