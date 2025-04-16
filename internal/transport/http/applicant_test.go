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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplicantHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		requestBody      dto.ApplicantRegister
		mocketup         func(applicant *mock.MockApplicant, auth *mock.MockAuth)
		expectedStatus   int
		expectedResponse any
	}{
		{
			name: "успешная регистрация",
			requestBody: dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "test@example.com",
					Password: "strongpassword",
				},
				FirstName: "Имя",
				LastName:  "Фамилия",
			},
			mocketup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
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
			name: "невалидные данные",
			requestBody: dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "invalid_email",
					Password: "strong!Password",
				},
			},
			mocketup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
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
		//{
		//	name: "невалидный JSON",
		//	requestBody: dto.ApplicantRegister{
		//		FirstName: "Имя",
		//	},
		//	mocketup:       func(applicant *mock.MockApplicant, auth *mock.MockAuth) {},
		//	expectedStatus: http.StatusBadRequest,
		//	expectedResponse: utils.APIError{
		//		Status:  http.StatusBadRequest,
		//		Message: "не удалось преобразовать данные в JSON",
		//	},
		//},
		{
			name: "пользователь уже существует",
			requestBody: dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "existing@example.com",
					Password: "somepassword",
				},
				FirstName: "Имя",
				LastName:  "Фамилия",
			},
			mocketup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
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
			requestBody: dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "user@example.com",
					Password: "somepassword",
				},
				FirstName: "Имя",
				LastName:  "Фамилия",
			},
			mocketup: func(applicant *mock.MockApplicant, auth *mock.MockAuth) {
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

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApplicant := mock.NewMockApplicant(ctrl)
			mockAuth := mock.NewMockAuth(ctrl)

			tt.mocketup(mockApplicant, mockAuth)

			cfg := config.CSRFConfig{
				CookieName: "csrf_token",
				Lifetime:   3600,
				Secret:     "secret",
				HttpOnly:   true,
				Secure:     false,
				SameSite:   "Strict",
			}
			handler := NewApplicantHandler(mockAuth, mockApplicant, nil, cfg)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/applicant/register", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.expectedStatus, res.StatusCode)

			if res.StatusCode == http.StatusOK {
				var success dto.AuthResponse
				err := json.NewDecoder(res.Body).Decode(&success)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResponse, success)
			} else {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, tt.expectedResponse, apiErr)
			}
		})
	}
}
