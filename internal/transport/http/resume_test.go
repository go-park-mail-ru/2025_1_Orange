package http

import (
	// "ResuMatch/internal/config"
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase/mock"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResumeHandler_CreateResume(t *testing.T) {
	t.Parallel()

	// Helper function to create valid resume request
	validResumeRequest := func() *dto.CreateResumeRequest {
		return &dto.CreateResumeRequest{
			AboutMe:                   "About me text",
			Specialization:            "Backend",
			Profession:                "Go Developer",
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
					EndDate:      "2021-01-01",
					UntilNow:     false,
				},
			},
		}
	}

	tests := []struct {
		name           string
		setupMock      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		cookie         *http.Cookie
		requestBody    interface{}
		expectedStatus int
		expectedError  error
	}{
		{
			name: "Success",
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Create(gomock.Any(), 1, gomock.Any()).Return(&dto.ResumeResponse{ID: 1}, nil)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody:    validResumeRequest(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "No cookie - unauthorized",
			setupMock:      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			cookie:         nil,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  entity.ErrUnauthorized,
		},
		{
			name: "Invalid session - unauthorized",
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid").Return(0, "", entity.ErrUnauthorized)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "invalid"},
			requestBody:    validResumeRequest(),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  entity.ErrInternal,
		},
		{
			name: "Not applicant - forbidden",
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "employer", nil)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody:    validResumeRequest(),
			expectedStatus: http.StatusForbidden,
			expectedError:  entity.ErrForbidden,
		},
		// {
		// 	name:           "Invalid JSON - bad request",
		// 	setupMock:      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
		// 	cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
		// 	requestBody:    "{invalid json}",
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedError:  entity.ErrBadRequest,
		// },
		{
			name: "Invalid request data - bad request",
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			cookie: &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: &dto.CreateResumeRequest{
				Specialization:         "", // invalid - required
				Profession:             "Go Developer",
				Education:              entity.Higher,
				EducationalInstitution: "University",
				GraduationYear:         "2020",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Create resume error - internal",
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Create(gomock.Any(), 1, gomock.Any()).Return(nil, entity.ErrInternal)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody:    validResumeRequest(),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  entity.ErrInternal,
		},
		{
			name: "JSON encode error - internal",
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Create(gomock.Any(), 1, gomock.Any()).Return(&dto.ResumeResponse{ID: 1}, nil)
			},
			cookie:         &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody:    validResumeRequest(),
			expectedStatus: http.StatusCreated,
			// Здесь мы не можем проверить ошибку кодирования JSON, так как httptest.ResponseRecorder не возвращает ошибки
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockResume := mock.NewMockResumeUsecase(ctrl)
			tt.setupMock(mockAuth, mockResume)

			handler := &ResumeHandler{
				auth:   mockAuth,
				resume: mockResume,
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

			req := httptest.NewRequest(http.MethodPost, "/resume/create", bytes.NewReader(body))
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			w := httptest.NewRecorder()

			handler.CreateResume(w, req)

			resp := w.Result()
			// defer resp.Body.Close()
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
				var resume dto.ResumeResponse
				err := json.NewDecoder(resp.Body).Decode(&resume)
				require.NoError(t, err)
				require.NotEmpty(t, resume.ID)
			}
		})
	}
}

func TestResumeHandler_GetResume(t *testing.T) {
	t.Parallel()

	// Helper function to create a valid resume response
	validResumeResponse := func() *dto.ResumeResponse {
		return &dto.ResumeResponse{
			ID:                        1,
			ApplicantID:               1,
			AboutMe:                   "About me",
			Specialization:            "Backend",
			Profession:                "Go Developer",
			Education:                 entity.Higher,
			EducationalInstitution:    "University",
			GraduationYear:            "2020",
			CreatedAt:                 "2023-01-01T00:00:00Z",
			UpdatedAt:                 "2023-01-02T00:00:00Z",
			Skills:                    []string{"Go", "SQL"},
			AdditionalSpecializations: []string{"DevOps"},
			WorkExperiences: []dto.WorkExperienceResponse{
				{
					ID:           1,
					EmployerName: "Company",
					Position:     "Developer",
					Duties:       "Development",
					Achievements: "Achievements",
					StartDate:    "2020-01-01",
					EndDate:      "2021-01-01",
					UntilNow:     false,
					UpdatedAt:    "2021-01-01T00:00:00Z",
				},
			},
		}
	}

	tests := []struct {
		name           string
		resumeID       string
		setupMock      func(resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedError  error
	}{
		{
			name:     "Success",
			resumeID: "1",
			setupMock: func(resume *mock.MockResumeUsecase) {
				resume.EXPECT().GetByID(gomock.Any(), 1).Return(validResumeResponse(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid resume ID - bad request",
			resumeID:       "invalid",
			setupMock:      func(resume *mock.MockResumeUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:     "Resume not found",
			resumeID: "1",
			setupMock: func(resume *mock.MockResumeUsecase) {
				resume.EXPECT().GetByID(gomock.Any(), 1).Return(nil, entity.NewError(entity.ErrNotFound, fmt.Errorf("resume not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  fmt.Errorf("resume not found"),
		},
		{
			name:     "Internal server error from usecase",
			resumeID: "1",
			setupMock: func(resume *mock.MockResumeUsecase) {
				resume.EXPECT().GetByID(gomock.Any(), 1).Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("database error")))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  fmt.Errorf("database error"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Initialize gomock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create mocks
			mockAuth := mock.NewMockAuth(ctrl)
			mockResume := mock.NewMockResumeUsecase(ctrl)
			tt.setupMock(mockResume)

			// Create handler
			handler := &ResumeHandler{
				auth:   mockAuth,
				resume: mockResume,
			}

			// Create HTTP request
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/resume/%s", tt.resumeID), nil)
			req = req.WithContext(req.Context())
			req.SetPathValue("id", tt.resumeID)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			handler.GetResume(w, req)

			// Verify response
			resp := w.Result()
			// defer resp.Body.Close()
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
				var resume dto.ResumeResponse
				err := json.NewDecoder(resp.Body).Decode(&resume)
				require.NoError(t, err)
				require.Equal(t, validResumeResponse().ID, resume.ID)
			}
		})
	}
}
func TestResumeHandler_UpdateResume(t *testing.T) {
	t.Parallel()

	// Helper function to create valid update resume request
	validUpdateResumeRequest := func() *dto.UpdateResumeRequest {
		return &dto.UpdateResumeRequest{
			AboutMe:                   "Updated about me text",
			Specialization:            "Backend",
			Profession:                "Senior Go Developer",
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
					EndDate:      "2021-01-01",
					UntilNow:     false,
				},
			},
		}
	}

	// Helper function to create valid resume response
	validResumeResponse := func() *dto.ResumeResponse {
		return &dto.ResumeResponse{
			ID:                        1,
			ApplicantID:               1,
			AboutMe:                   "Updated about me text",
			Specialization:            "Backend",
			Profession:                "Senior Go Developer",
			Education:                 entity.Higher,
			EducationalInstitution:    "University",
			GraduationYear:            "2020",
			CreatedAt:                 "2023-01-01T00:00:00Z",
			UpdatedAt:                 "2023-01-02T00:00:00Z",
			Skills:                    []string{"Go", "SQL"},
			AdditionalSpecializations: []string{"DevOps"},
			WorkExperiences: []dto.WorkExperienceResponse{
				{
					ID:           1,
					EmployerName: "Company",
					Position:     "Developer",
					Duties:       "Development",
					Achievements: "Achievements",
					StartDate:    "2020-01-01",
					EndDate:      "2021-01-01",
					UntilNow:     false,
					UpdatedAt:    "2021-01-01T00:00:00Z",
				},
			},
		}
	}

	tests := []struct {
		name           string
		resumeID       string
		cookie         *http.Cookie
		requestBody    interface{}
		setupMock      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest)
		expectedStatus int
		expectedError  error
	}{
		{
			name:        "Success",
			resumeID:    "1",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: validUpdateResumeRequest(),
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Update(gomock.Any(), 1, 1, gomock.Any()).Return(validResumeResponse(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No cookie - unauthorized",
			resumeID:       "1",
			cookie:         nil,
			requestBody:    validUpdateResumeRequest(),
			setupMock:      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  entity.ErrUnauthorized,
		},
		{
			name:        "Invalid session - unauthorized",
			resumeID:    "1",
			cookie:      &http.Cookie{Name: "session_id", Value: "invalid"},
			requestBody: validUpdateResumeRequest(),
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid").Return(0, "", entity.NewError(entity.ErrUnauthorized, fmt.Errorf("invalid session")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  fmt.Errorf("invalid session"),
		},
		{
			name:        "Not applicant - forbidden",
			resumeID:    "1",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: validUpdateResumeRequest(),
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  entity.ErrForbidden,
		},
		{
			name:        "Invalid resume ID - bad request",
			resumeID:    "invalid",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: validUpdateResumeRequest(),
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:        "Invalid JSON - bad request",
			resumeID:    "1",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: `{invalid json}`,
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:     "Invalid request data - bad request",
			resumeID: "1",
			cookie:   &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: &dto.UpdateResumeRequest{
				Specialization:         "", // invalid - required
				Profession:             "Go Developer",
				Education:              entity.Higher,
				EducationalInstitution: "University",
				GraduationYear:         "2020",
			},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "Resume not found",
			resumeID:    "1",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: validUpdateResumeRequest(),
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Update(gomock.Any(), 1, 1, gomock.Any()).Return(nil, entity.NewError(entity.ErrNotFound, fmt.Errorf("resume not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  fmt.Errorf("resume not found"),
		},
		{
			name:        "Not owner - forbidden",
			resumeID:    "1",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: validUpdateResumeRequest(),
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Update(gomock.Any(), 1, 1, gomock.Any()).Return(nil, entity.NewError(entity.ErrForbidden, fmt.Errorf("not authorized to update resume")))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  fmt.Errorf("not authorized to update resume"),
		},
		{
			name:        "Internal server error",
			resumeID:    "1",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			requestBody: validUpdateResumeRequest(),
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase, request *dto.UpdateResumeRequest) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Update(gomock.Any(), 1, 1, gomock.Any()).Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("database error")))
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
			mockResume := mock.NewMockResumeUsecase(ctrl)

			// Create request body
			var body []byte
			switch v := tt.requestBody.(type) {
			case string:
				body = []byte(v)
			default:
				var err error
				body, err = json.Marshal(v)
				require.NoError(t, err)
			}

			// Prepare sanitized request for mock expectation
			var sanitizedRequest *dto.UpdateResumeRequest
			if req, ok := tt.requestBody.(*dto.UpdateResumeRequest); ok {
				sanitizedRequest = &dto.UpdateResumeRequest{
					AboutMe:                   req.AboutMe,
					Specialization:            req.Specialization,
					Profession:                req.Profession,
					Education:                 req.Education,
					EducationalInstitution:    req.EducationalInstitution,
					GraduationYear:            req.GraduationYear,
					Skills:                    make([]string, len(req.Skills)),
					AdditionalSpecializations: make([]string, len(req.AdditionalSpecializations)),
					WorkExperiences:           make([]dto.WorkExperienceDTO, len(req.WorkExperiences)),
				}
				copy(sanitizedRequest.Skills, req.Skills)
				copy(sanitizedRequest.AdditionalSpecializations, req.AdditionalSpecializations)
				copy(sanitizedRequest.WorkExperiences, req.WorkExperiences)
			}

			tt.setupMock(mockAuth, mockResume, sanitizedRequest)

			handler := &ResumeHandler{
				auth:   mockAuth,
				resume: mockResume,
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/resume/%s", tt.resumeID), bytes.NewReader(body))
			req = req.WithContext(req.Context())
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			req.SetPathValue("id", tt.resumeID)

			w := httptest.NewRecorder()

			handler.UpdateResume(w, req)

			resp := w.Result()
			// defer resp.Body.Close()
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
				if tt.expectedStatus == http.StatusBadRequest {
					// For bad request, the error message may include validation details
					require.Contains(t, apiError.Message, tt.expectedError.Error())
				} else {
					require.Equal(t, tt.expectedError.Error(), apiError.Message)
				}
				require.Equal(t, tt.expectedStatus, apiError.Status)
			} else if resp.StatusCode == http.StatusOK {
				var resume dto.ResumeResponse
				err := json.NewDecoder(resp.Body).Decode(&resume)
				require.NoError(t, err)
				require.Equal(t, validResumeResponse().ID, resume.ID)
			}
		})
	}
}

func TestResumeHandler_DeleteResume(t *testing.T) {
	t.Parallel()

	// Helper function to create valid delete resume response
	validDeleteResumeResponse := func() *dto.DeleteResumeResponse {
		return &dto.DeleteResumeResponse{
			Success: true,
			Message: "Resume deleted successfully",
		}
	}

	tests := []struct {
		name           string
		resumeID       string
		cookie         *http.Cookie
		setupMock      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedError  error
	}{
		{
			name:     "Success",
			resumeID: "1",
			cookie:   &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Delete(gomock.Any(), 1, 1).Return(validDeleteResumeResponse(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No cookie - unauthorized",
			resumeID:       "1",
			cookie:         nil,
			setupMock:      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  entity.ErrUnauthorized,
		},
		{
			name:     "Invalid session - unauthorized",
			resumeID: "1",
			cookie:   &http.Cookie{Name: "session_id", Value: "invalid"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid").Return(0, "", entity.NewError(entity.ErrUnauthorized, fmt.Errorf("invalid session")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  fmt.Errorf("invalid session"),
		},
		{
			name:     "Not applicant - forbidden",
			resumeID: "1",
			cookie:   &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "employer", nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  entity.ErrForbidden,
		},
		{
			name:     "Invalid resume ID - bad request",
			resumeID: "invalid",
			cookie:   &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:     "Resume not found",
			resumeID: "1",
			cookie:   &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Delete(gomock.Any(), 1, 1).Return(nil, entity.NewError(entity.ErrNotFound, fmt.Errorf("resume not found")))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  fmt.Errorf("resume not found"),
		},
		{
			name:     "Not owner - forbidden",
			resumeID: "1",
			cookie:   &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Delete(gomock.Any(), 1, 1).Return(nil, entity.NewError(entity.ErrForbidden, fmt.Errorf("not authorized to delete resume")))
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  fmt.Errorf("not authorized to delete resume"),
		},
		{
			name:     "Internal server error",
			resumeID: "1",
			cookie:   &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().Delete(gomock.Any(), 1, 1).Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("database error")))
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
			mockResume := mock.NewMockResumeUsecase(ctrl)
			tt.setupMock(mockAuth, mockResume)

			handler := &ResumeHandler{
				auth:   mockAuth,
				resume: mockResume,
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/resume/%s", tt.resumeID), nil)
			req = req.WithContext(req.Context())
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			req.SetPathValue("id", tt.resumeID)

			w := httptest.NewRecorder()

			handler.DeleteResume(w, req)

			resp := w.Result()
			// defer resp.Body.Close()
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
				var response dto.DeleteResumeResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.True(t, response.Success)
				require.Equal(t, validDeleteResumeResponse().Message, response.Message)
			}
		})
	}
}
func TestResumeHandler_GetAllResumes(t *testing.T) {
	t.Parallel()

	// Helper function to create valid resume short response (non-applicant)
	validResumeShortResponse := func() []dto.ResumeShortResponse {
		return []dto.ResumeShortResponse{
			{
				ID:             1,
				ApplicantID:    1,
				Specialization: "Backend",
				Profession:     "Go Developer",
				CreatedAt:      "2023-01-01T00:00:00Z",
				UpdatedAt:      "2023-01-02T00:00:00Z",
			},
		}
	}

	// Helper function to create valid resume applicant short response (applicant)
	validResumeApplicantShortResponse := func() []dto.ResumeApplicantShortResponse {
		return []dto.ResumeApplicantShortResponse{
			{
				ID:             1,
				ApplicantID:    1,
				Specialization: "Backend",
				Profession:     "Go Developer",
				Skills:         []string{"Go", "SQL"},
				CreatedAt:      "2023-01-01T00:00:00Z",
				UpdatedAt:      "2023-01-02T00:00:00Z",
			},
		}
	}

	tests := []struct {
		name           string
		queryParams    string
		cookie         *http.Cookie
		setupMock      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedError  error
		isApplicant    bool
	}{
		{
			name:        "Success - applicant",
			queryParams: "limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).Return(validResumeApplicantShortResponse(), nil)
			},
			expectedStatus: http.StatusOK,
			isApplicant:    true,
		},
		{
			name:        "Success - non-applicant",
			queryParams: "limit=20&offset=10",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "employer", nil)
				resume.EXPECT().GetAll(gomock.Any(), 20, 10).Return(validResumeShortResponse(), nil)
			},
			expectedStatus: http.StatusOK,
			isApplicant:    false,
		},
		{
			name:           "No cookie - unauthorized",
			queryParams:    "limit=10&offset=0",
			cookie:         nil,
			setupMock:      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  entity.ErrUnauthorized,
		},
		{
			name:        "Invalid session - unauthorized",
			queryParams: "limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "invalid"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid").Return(0, "", entity.NewError(entity.ErrUnauthorized, fmt.Errorf("invalid session")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  fmt.Errorf("invalid session"),
		},
		{
			name:        "Invalid limit - bad request",
			queryParams: "limit=invalid&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:        "Invalid offset - bad request",
			queryParams: "limit=10&offset=invalid",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:        "Internal server error - applicant",
			queryParams: "limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("database error")))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  fmt.Errorf("database error"),
			isApplicant:    true,
		},
		{
			name:        "Internal server error - non-applicant",
			queryParams: "limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "employer", nil)
				resume.EXPECT().GetAll(gomock.Any(), 10, 0).Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("database error")))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  fmt.Errorf("database error"),
			isApplicant:    false,
		},
		{
			name:        "Default pagination - applicant",
			queryParams: "",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).Return(validResumeApplicantShortResponse(), nil)
			},
			expectedStatus: http.StatusOK,
			isApplicant:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockResume := mock.NewMockResumeUsecase(ctrl)
			tt.setupMock(mockAuth, mockResume)

			handler := &ResumeHandler{
				auth:   mockAuth,
				resume: mockResume,
			}

			url := "/resume/all"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = req.WithContext(req.Context())
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			w := httptest.NewRecorder()

			handler.GetAllResumes(w, req)

			resp := w.Result()
			// defer resp.Body.Close()
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
				if tt.isApplicant {
					var resumes []dto.ResumeApplicantShortResponse
					err := json.NewDecoder(resp.Body).Decode(&resumes)
					require.NoError(t, err)
					require.Len(t, resumes, len(validResumeApplicantShortResponse()))
					require.Equal(t, validResumeApplicantShortResponse()[0].ID, resumes[0].ID)
				} else {
					var resumes []dto.ResumeShortResponse
					err := json.NewDecoder(resp.Body).Decode(&resumes)
					require.NoError(t, err)
					require.Len(t, resumes, len(validResumeShortResponse()))
					require.Equal(t, validResumeShortResponse()[0].ID, resumes[0].ID)
				}
			}
		})
	}
}
func TestResumeHandler_SearchResumes(t *testing.T) {
	t.Parallel()

	// Helper function to create valid resume short response
	validResumeShortResponse := func() []dto.ResumeShortResponse {
		return []dto.ResumeShortResponse{
			{
				ID:             1,
				ApplicantID:    1,
				Specialization: "Backend",
				Profession:     "Go Developer",
				CreatedAt:      "2023-01-01T00:00:00Z",
				UpdatedAt:      "2023-01-02T00:00:00Z",
			},
		}
	}

	tests := []struct {
		name           string
		queryParams    string
		cookie         *http.Cookie
		setupMock      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase)
		expectedStatus int
		expectedError  error
		role           string
	}{
		{
			name:        "Success - applicant",
			queryParams: "profession=Go+Developer&limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().SearchResumesByProfession(gomock.Any(), 1, "applicant", "Go Developer", 10, 0).Return(validResumeShortResponse(), nil)
			},
			expectedStatus: http.StatusOK,
			role:           "applicant",
		},
		{
			name:        "Success - non-applicant",
			queryParams: "profession=Go+Developer&limit=20&offset=10",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "employer", nil)
				resume.EXPECT().SearchResumesByProfession(gomock.Any(), 1, "employer", "Go Developer", 20, 10).Return(validResumeShortResponse(), nil)
			},
			expectedStatus: http.StatusOK,
			role:           "employer",
		},
		{
			name:           "No cookie - unauthorized",
			queryParams:    "profession=Go+Developer&limit=10&offset=0",
			cookie:         nil,
			setupMock:      func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  entity.ErrUnauthorized,
		},
		{
			name:        "Invalid session - unauthorized",
			queryParams: "profession=Go+Developer&limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "invalid"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "invalid").Return(0, "", entity.NewError(entity.ErrUnauthorized, fmt.Errorf("invalid session")))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  fmt.Errorf("invalid session"),
		},
		{
			name:        "Missing profession - bad request",
			queryParams: "limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  fmt.Errorf("параметр profession обязателен"),
		},
		{
			name:        "Invalid limit - bad request",
			queryParams: "profession=Go+Developer&limit=invalid&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:        "Invalid offset - bad request",
			queryParams: "profession=Go+Developer&limit=10&offset=invalid",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  entity.ErrBadRequest,
		},
		{
			name:        "Internal server error",
			queryParams: "profession=Go+Developer&limit=10&offset=0",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().SearchResumesByProfession(gomock.Any(), 1, "applicant", "Go Developer", 10, 0).Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("database error")))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  fmt.Errorf("database error"),
			role:           "applicant",
		},
		{
			name:        "Default pagination",
			queryParams: "profession=Go+Developer",
			cookie:      &http.Cookie{Name: "session_id", Value: "session123"},
			setupMock: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
				auth.EXPECT().GetUserIDBySession(gomock.Any(), "session123").Return(1, "applicant", nil)
				resume.EXPECT().SearchResumesByProfession(gomock.Any(), 1, "applicant", "Go Developer", 10, 0).Return(validResumeShortResponse(), nil)
			},
			expectedStatus: http.StatusOK,
			role:           "applicant",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuth := mock.NewMockAuth(ctrl)
			mockResume := mock.NewMockResumeUsecase(ctrl)
			tt.setupMock(mockAuth, mockResume)

			handler := &ResumeHandler{
				auth:   mockAuth,
				resume: mockResume,
			}

			url := "/resume/search?" + tt.queryParams
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req = req.WithContext(req.Context())
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			w := httptest.NewRecorder()

			handler.SearchResumes(w, req)

			resp := w.Result()
			// defer resp.Body.Close()
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
				var resumes []dto.ResumeShortResponse
				err := json.NewDecoder(resp.Body).Decode(&resumes)
				require.NoError(t, err)
				require.Len(t, resumes, len(validResumeShortResponse()))
				require.Equal(t, validResumeShortResponse()[0].ID, resumes[0].ID)
			}
		})
	}
}
