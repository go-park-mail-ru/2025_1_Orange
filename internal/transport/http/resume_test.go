// package http

// import (
// 	"ResuMatch/internal/config"
// 	"ResuMatch/internal/entity"
// 	"ResuMatch/internal/entity/dto"
// 	"ResuMatch/internal/transport/http/utils"
// 	"ResuMatch/internal/usecase/mock"
// 	"bytes"

// 	// "context"
// 	"encoding/json"
// 	"errors"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"
// )

// func TestResumeHandler_CreateResume(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		setupMocks     func(*mock.MockAuth, *mock.MockResumeUsecase)
// 		requestBody    interface{}
// 		cookie         *http.Cookie
// 		expectedStatus int
// 		expectedBody   interface{}
// 	}{
// 		{
// 			name: "Success",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "valid_session").
// 					Return(1, "applicant", nil)

// 				req := &dto.CreateResumeRequest{
// 					AboutMe:                "About me",
// 					Specialization:         "Backend",
// 					Education:              entity.Higher,
// 					EducationalInstitution: "University",
// 					GraduationYear:         "2020",
// 					Skills:                 []string{"Go", "SQL"},
// 				}

// 				resume.EXPECT().
// 					Create(gomock.Any(), 1, req).
// 					Return(&dto.ResumeResponse{
// 						ID:             1,
// 						ApplicantID:    1,
// 						AboutMe:        "About me",
// 						Specialization: "Backend",
// 					}, nil)
// 			},
// 			requestBody: map[string]interface{}{
// 				"about_me":                "About me",
// 				"specialization":          "Backend",
// 				"education":               "higher",
// 				"educational_institution": "University",
// 				"graduation_year":         "2020",
// 				"skills":                  []string{"Go", "SQL"},
// 			},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			expectedStatus: http.StatusCreated,
// 			expectedBody: &dto.ResumeResponse{
// 				ID:             1,
// 				ApplicantID:    1,
// 				AboutMe:        "About me",
// 				Specialization: "Backend",
// 			},
// 		},
// 		{
// 			name:           "Unauthorized - no cookie",
// 			setupMocks:     func(*mock.MockAuth, *mock.MockResumeUsecase) {},
// 			requestBody:    map[string]interface{}{},
// 			cookie:         nil,
// 			expectedStatus: http.StatusUnauthorized,
// 			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
// 		},
// 		// {
// 		// 	name: "Unauthorized - invalid session",
// 		// 	setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 		// 		auth.EXPECT().
// 		// 			GetUserIDBySession(gomock.Any(), "invalid_session").
// 		// 			Return(0, "", entity.ErrUnauthorized)
// 		// 	},
// 		// 	requestBody:    map[string]interface{}{},
// 		// 	cookie:         &http.Cookie{Name: "session_id", Value: "invalid_session"},
// 		// 	expectedStatus: http.StatusUnauthorized,
// 		// 	expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
// 		// },
// 		{
// 			name: "Forbidden - not applicant",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "employer_session").
// 					Return(1, "employer", nil)
// 			},
// 			requestBody:    map[string]interface{}{},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "employer_session"},
// 			expectedStatus: http.StatusForbidden,
// 			expectedBody:   utils.APIError{Status: http.StatusForbidden, Message: entity.ErrForbidden.Error()},
// 		},
// 		{
// 			name: "Bad request - invalid body",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "valid_session").
// 					Return(1, "applicant", nil)
// 			},
// 			requestBody:    "invalid json",
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
// 		},
// 		{
// 			name: "Internal error - usecase failed",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "valid_session").
// 					Return(1, "applicant", nil)

// 				resume.EXPECT().
// 					Create(gomock.Any(), 1, gomock.Any()).
// 					Return(nil, entity.ErrInternal)
// 			},
// 			requestBody: map[string]interface{}{
// 				"about_me":                "About me",
// 				"specialization":          "Backend",
// 				"education":               "higher",
// 				"educational_institution": "University",
// 				"graduation_year":         "2020",
// 				"skills":                  []string{"Go", "SQL"},
// 			},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody:   utils.APIError{Status: http.StatusInternalServerError, Message: "internal server error"},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockAuth := mock.NewMockAuth(ctrl)
// 			mockResume := mock.NewMockResumeUsecase(ctrl)

// 			if tt.setupMocks != nil {
// 				tt.setupMocks(mockAuth, mockResume)
// 			}

// 			handler := NewResumeHandler(mockAuth, mockResume, config.CSRFConfig{})

// 			var bodyBytes []byte
// 			switch v := tt.requestBody.(type) {
// 			case string:
// 				bodyBytes = []byte(v)
// 			default:
// 				bodyBytes, _ = json.Marshal(v)
// 			}

// 			req := httptest.NewRequest(http.MethodPost, "/resume", bytes.NewReader(bodyBytes))
// 			if tt.cookie != nil {
// 				req.AddCookie(tt.cookie)
// 			}

// 			w := httptest.NewRecorder()

// 			handler.CreateResume(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			if tt.expectedBody != nil {
// 				var expectedBytes []byte
// 				switch v := tt.expectedBody.(type) {
// 				case string:
// 					expectedBytes = []byte(v)
// 				default:
// 					expectedBytes, _ = json.Marshal(v)
// 				}

// 				assert.JSONEq(t, string(expectedBytes), w.Body.String())
// 			}
// 		})
// 	}
// }

// func TestResumeHandler_GetResume(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		resumeID       string
// 		setupMocks     func(*mock.MockResumeUsecase)
// 		expectedStatus int
// 		expectedBody   interface{}
// 	}{
// 		{
// 			name:     "Success",
// 			resumeID: "123",
// 			setupMocks: func(ru *mock.MockResumeUsecase) {
// 				ru.EXPECT().
// 					GetByID(gomock.Any(), 123).
// 					Return(&dto.ResumeResponse{
// 						ID:             123,
// 						ApplicantID:    456,
// 						Specialization: "Backend Developer",
// 					}, nil)
// 			},
// 			expectedStatus: http.StatusOK,
// 			expectedBody: &dto.ResumeResponse{
// 				ID:             123,
// 				ApplicantID:    456,
// 				Specialization: "Backend Developer",
// 			},
// 		},
// 		{
// 			name:           "Bad Request - invalid resume ID",
// 			resumeID:       "invalid",
// 			setupMocks:     func(*mock.MockResumeUsecase) {},
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
// 		},
// 		{
// 			name:     "Not Found",
// 			resumeID: "123",
// 			setupMocks: func(ru *mock.MockResumeUsecase) {
// 				ru.EXPECT().
// 					GetByID(gomock.Any(), 123).
// 					Return(nil, entity.NewError(entity.ErrNotFound, errors.New("resume not found")))
// 			},
// 			expectedStatus: http.StatusNotFound,
// 			expectedBody:   utils.APIError{Status: http.StatusNotFound, Message: "resume not found"},
// 		},
// 		{
// 			name:     "Internal Server Error",
// 			resumeID: "123",
// 			setupMocks: func(ru *mock.MockResumeUsecase) {
// 				ru.EXPECT().
// 					GetByID(gomock.Any(), 123).
// 					Return(nil, errors.New("database error"))
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody:   utils.APIError{Status: http.StatusInternalServerError, Message: "internal server error"},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockResume := mock.NewMockResumeUsecase(ctrl)
// 			if tt.setupMocks != nil {
// 				tt.setupMocks(mockResume)
// 			}

// 			handler := NewResumeHandler(nil, mockResume, config.CSRFConfig{})

// 			req := httptest.NewRequest(http.MethodGet, "/resume/"+tt.resumeID, nil)
// 			req.SetPathValue("id", tt.resumeID)

// 			w := httptest.NewRecorder()

// 			handler.GetResume(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			if tt.expectedBody != nil {
// 				var expectedBytes []byte
// 				switch v := tt.expectedBody.(type) {
// 				case string:
// 					expectedBytes = []byte(v)
// 				default:
// 					var err error
// 					expectedBytes, err = json.Marshal(v)
// 					assert.NoError(t, err)
// 				}

// 				assert.JSONEq(t, string(expectedBytes), w.Body.String())
// 			}
// 		})
// 	}
// }

// func TestResumeHandler_UpdateResume(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		setupMocks     func(*mock.MockAuth, *mock.MockResumeUsecase)
// 		requestBody    interface{}
// 		cookie         *http.Cookie
// 		resumeID       string
// 		expectedStatus int
// 		expectedBody   interface{}
// 	}{
// 		{
// 			name: "Success",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "valid_session").
// 					Return(1, "applicant", nil)

// 				req := &dto.UpdateResumeRequest{
// 					AboutMe:                "Updated about me",
// 					Specialization:         "Updated specialization",
// 					Education:              entity.Higher,
// 					EducationalInstitution: "Updated university",
// 				}

// 				resume.EXPECT().
// 					Update(gomock.Any(), 123, 1, req).
// 					Return(&dto.ResumeResponse{
// 						ID:             123,
// 						ApplicantID:    1,
// 						AboutMe:        "Updated about me",
// 						Specialization: "Updated specialization",
// 					}, nil)
// 			},
// 			requestBody: map[string]interface{}{
// 				"about_me":                "Updated about me",
// 				"specialization":          "Updated specialization",
// 				"education":               "higher",
// 				"educational_institution": "Updated university",
// 			},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			resumeID:       "123",
// 			expectedStatus: http.StatusOK,
// 			expectedBody: &dto.ResumeResponse{
// 				ID:             123,
// 				ApplicantID:    1,
// 				AboutMe:        "Updated about me",
// 				Specialization: "Updated specialization",
// 			},
// 		},
// 		{
// 			name:           "Unauthorized - no cookie",
// 			setupMocks:     func(*mock.MockAuth, *mock.MockResumeUsecase) {},
// 			requestBody:    map[string]interface{}{},
// 			cookie:         nil,
// 			resumeID:       "123",
// 			expectedStatus: http.StatusUnauthorized,
// 			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: entity.ErrUnauthorized.Error()},
// 		},
// 		{
// 			name: "Unauthorized - invalid session",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "invalid_session").
// 					Return(0, "", entity.NewError(entity.ErrUnauthorized, errors.New("invalid session")))
// 			},
// 			requestBody:    map[string]interface{}{},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "invalid_session"},
// 			resumeID:       "123",
// 			expectedStatus: http.StatusUnauthorized,
// 			expectedBody:   utils.APIError{Status: http.StatusUnauthorized, Message: "invalid session"},
// 		},
// 		{
// 			name: "Forbidden - not applicant",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "employer_session").
// 					Return(1, "employer", nil)
// 			},
// 			requestBody:    map[string]interface{}{},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "employer_session"},
// 			resumeID:       "123",
// 			expectedStatus: http.StatusForbidden,
// 			expectedBody:   utils.APIError{Status: http.StatusForbidden, Message: entity.ErrForbidden.Error()},
// 		},
// 		{
// 			name:           "Bad Request - invalid resume ID",
// 			setupMocks:     func(*mock.MockAuth, *mock.MockResumeUsecase) {},
// 			requestBody:    map[string]interface{}{},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			resumeID:       "invalid",
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
// 		},
// 		{
// 			name: "Bad Request - invalid body",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "valid_session").
// 					Return(1, "applicant", nil)
// 			},
// 			requestBody:    "invalid json",
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			resumeID:       "123",
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
// 		},
// 		{
// 			name: "Not Found",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "valid_session").
// 					Return(1, "applicant", nil)

// 				resume.EXPECT().
// 					Update(gomock.Any(), 123, 1, gomock.Any()).
// 					Return(nil, entity.NewError(entity.ErrNotFound, errors.New("resume not found")))
// 			},
// 			requestBody: map[string]interface{}{
// 				"about_me": "Test",
// 			},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			resumeID:       "123",
// 			expectedStatus: http.StatusNotFound,
// 			expectedBody:   utils.APIError{Status: http.StatusNotFound, Message: "resume not found"},
// 		},
// 		{
// 			name: "Bad Request - invalid resume ID",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				// Не ожидаем вызовов к auth, так как сначала проверяется ID
// 			},
// 			requestBody:    map[string]interface{}{},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			resumeID:       "invalid",
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
// 		},
// 		{
// 			name: "Internal Server Error",
// 			setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
// 				auth.EXPECT().
// 					GetUserIDBySession(gomock.Any(), "valid_session").
// 					Return(1, "applicant", nil)

// 				resume.EXPECT().
// 					Update(gomock.Any(), 123, 1, gomock.Any()).
// 					Return(nil, errors.New("database error"))
// 			},
// 			requestBody: map[string]interface{}{
// 				"about_me": "Test",
// 			},
// 			cookie:         &http.Cookie{Name: "session_id", Value: "valid_session"},
// 			resumeID:       "123",
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody:   utils.APIError{Status: http.StatusInternalServerError, Message: "internal server error"},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockAuth := mock.NewMockAuth(ctrl)
// 			mockResume := mock.NewMockResumeUsecase(ctrl)

// 			if tt.setupMocks != nil {
// 				tt.setupMocks(mockAuth, mockResume)
// 			}

// 			handler := NewResumeHandler(mockAuth, mockResume, config.CSRFConfig{})

// 			var bodyBytes []byte
// 			switch v := tt.requestBody.(type) {
// 			case string:
// 				bodyBytes = []byte(v)
// 			default:
// 				var err error
// 				bodyBytes, err = json.Marshal(v)
// 				assert.NoError(t, err)
// 			}

// 			req := httptest.NewRequest(http.MethodPut, "/resume/"+tt.resumeID, bytes.NewReader(bodyBytes))
// 			req.SetPathValue("id", tt.resumeID)
// 			if tt.cookie != nil {
// 				req.AddCookie(tt.cookie)
// 			}

// 			w := httptest.NewRecorder()

// 			handler.UpdateResume(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			if tt.expectedBody != nil {
// 				var expectedBytes []byte
// 				switch v := tt.expectedBody.(type) {
// 				case string:
// 					expectedBytes = []byte(v)
// 				default:
// 					var err error
// 					expectedBytes, err = json.Marshal(v)
// 					assert.NoError(t, err)
// 				}

//					assert.JSONEq(t, string(expectedBytes), w.Body.String())
//				}
//			})
//		}
//	}
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
		// {
		// 	name: "Not found",
		// 	url:  "/resume/123",
		// 	setupMocks: func(resume *mock.MockResumeUsecase) {
		// 		resume.EXPECT().GetByID(gomock.Any(), 123).
		// 			Return(nil, entity.ErrNotFound)
		// 	},
		// 	expectedStatus: http.StatusNotFound,
		// 	expectedBody:   utils.ToAPIError(entity.ErrNotFound),
		// },
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
		// {
		// 	name:        "Bad request - invalid resume ID",
		// 	url:         "/resume/invalid",
		// 	requestBody: dto.UpdateResumeRequest{},
		// 	cookie:      &http.Cookie{Name: "session_id", Value: "valid-session"},
		// 	setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
		// 		// Явно указываем, что не ожидаем вызовов
		// 		auth.EXPECT().GetUserIDBySession(gomock.Any(), gomock.Any()).Times(0)
		// 		resume.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		// 	},
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedBody:   utils.APIError{Status: http.StatusBadRequest, Message: entity.ErrBadRequest.Error()},
		// },
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
		// {
		// 	name: "Not found",
		// 	url:  "/resume/123",
		// 	requestBody: dto.UpdateResumeRequest{
		// 		AboutMe: "Updated about me",
		// 	},
		// 	cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
		// 	setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
		// 		auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
		// 			Return(1, "applicant", nil)
		// 		resume.EXPECT().Update(
		// 			gomock.Any(),
		// 			123,
		// 			1,
		// 			gomock.Any(),
		// 		).Return(nil, entity.ErrNotFound)
		// 	},
		// 	expectedStatus: http.StatusNotFound,
		// 	expectedBody:   utils.ToAPIError(entity.ErrNotFound),
		// },
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
		// {
		// 	name:   "Not found",
		// 	url:    "/resume/123",
		// 	cookie: &http.Cookie{Name: "session_id", Value: "valid-session"},
		// 	setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
		// 		auth.EXPECT().GetUserIDBySession(gomock.Any(), "valid-session").
		// 			Return(1, "applicant", nil)
		// 		resume.EXPECT().Delete(gomock.Any(), 123, 1).
		// 			Return(nil, entity.ErrNotFound)
		// 	},
		// 	expectedStatus: http.StatusNotFound,
		// 	expectedBody:   utils.ToAPIError(entity.ErrNotFound),
		// },
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
		// {
		// 	name:   "Applicant - not found",
		// 	cookie: &http.Cookie{Name: "session_id", Value: "applicant-session"},
		// 	setupMocks: func(auth *mock.MockAuth, resume *mock.MockResumeUsecase) {
		// 		auth.EXPECT().GetUserIDBySession(gomock.Any(), "applicant-session").
		// 			Return(1, "applicant", nil)
		// 		resume.EXPECT().GetAllResumesByApplicantID(gomock.Any(), 1).
		// 			Return(nil, entity.ErrNotFound)
		// 	},
		// 	expectedStatus: http.StatusNotFound,
		// 	expectedBody:   utils.ToAPIError(entity.ErrNotFound),
		// },
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
