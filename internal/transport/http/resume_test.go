package http

import (
	"ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase/mock"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResumeHandler_CreateResume(t *testing.T) {
	t.Parallel()

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
				Profession:                "Developer",
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
		// {
		// 	name: "Unauthorized - invalid session",
		// 	setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
		// 		auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid-session").
		// 			Return(0, "", entity.ErrUnauthorized)
		// 	},
		// 	requestBody:    dto.CreateResumeRequest{},
		// 	cookie:         &http.Cookie{Name: "session_id", Value: "invalid-session"},
		// 	expectedStatus: http.StatusUnauthorized,
		// 	expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
		// },
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
				Profession:     "Developer",
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "valid-session"},
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
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			tc.setupMocks(authMock, resumeMock)

			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

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

			rec := httptest.NewRecorder()

			handler.CreateResume(rec, req)

			require.Equal(t, tc.expectedStatus, rec.Code)

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
		url            string
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
						Profession:     "Developer",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &dto.ResumeResponse{
				ID:             123,
				Specialization: "Backend",
				Profession:     "Developer",
			},
		},
		{
			name: "Bad request - invalid resume ID",
			url:  "/resume/invalid",
			setupMocks: func(resume *mock.MockResumeUsecase) {
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
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			tc.setupMocks(resumeMock)

			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

			req, err := http.NewRequest(http.MethodGet, tc.url, nil)
			require.NoError(t, err)

			router := http.NewServeMux()
			handler.Configure(router)

			rec := httptest.NewRecorder()

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
				require.Equal(t, tc.expectedBody, resp)
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
				Profession:     "Updated profession",
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
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
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
				AboutMe:    "Updated about me",
				Profession: "Developer",
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
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			tc.setupMocks(authMock, resumeMock)

			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

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

			router := http.NewServeMux()
			handler.Configure(router)

			rec := httptest.NewRecorder()

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
				require.Equal(t, tc.expectedBody, resp)
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
		// {
		// 	name:   "Unauthorized - invalid session",
		// 	url:    "/resume/123",
		// 	cookie: &http.Cookie{Name: "session_id", Value: "invalid-session"},
		// 	setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
		// 		auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid-session").
		// 			Return(0, "", entity.ErrUnauthorized)
		// 	},
		// 	expectedStatus: http.StatusUnauthorized,
		// 	expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
		// },
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
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
					Return(1, "applicant", nil)
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
				require.Equal(t, tc.expectedBody, resp)
			}
		})
	}
}

func TestResumeHandler_GetAllResumes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		queryParams    string
		cookie         *http.Cookie
		setupMocks     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "Success - applicant role",
			queryParams: "?limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "applicant-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "applicant-session").
					Return(1, "applicant", nil)
				resume.EXPECT().GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
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
			name:        "Success - employer role",
			queryParams: "?limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "employer-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "employer-session").
					Return(2, "employer", nil)
				resume.EXPECT().GetAll(gomock.Any(), 10, 0).
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
			queryParams:    "",
			cookie:         nil,
			setupMocks:     func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
		},
		{
			name:        "Unauthorized - invalid session",
			queryParams: "",
			cookie:      &http.Cookie{Name: "session_id", Value: "invalid-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid-session").
					Return(0, "", entity.ErrInternal)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   utils.APIError{Status: http.StatusInternalServerError, Message: entity.ErrInternal.Error()},
		},
		{
			name:        "Employer - internal error",
			queryParams: "?limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "employer-session"},
			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "employer-session").
					Return(2, "employer", nil)
				resume.EXPECT().GetAll(gomock.Any(), 10, 0).
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
			resumeMock := mock.NewMockResumeUsecase(ctrl)

			tc.setupMocks(authMock, resumeMock)

			handler := NewResumeHandler(authMock, resumeMock, config.CSRFConfig{})

			req := httptest.NewRequest(http.MethodGet, "/resume/all"+tc.queryParams, nil)
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
				require.Equal(t, tc.expectedBody, resp)
			}
		})
	}
}
