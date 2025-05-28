package http

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/transport/http/utils"
	"ResuMatch/internal/usecase/mock"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSpecializationHandler_GetAllSpecializationNames(t *testing.T) {
	t.Parallel()

	// Helper function to create valid specialization names response
	validSpecializationNamesResponse := func() *dto.SpecializationNamesResponse {
		return &dto.SpecializationNamesResponse{
			Names: []string{"Backend", "Frontend", "DevOps"},
		}
	}

	tests := []struct {
		name           string
		setupMock      func(specialization *mock.MockSpecializationUsecase)
		expectedStatus int
		expectedError  error
	}{
		{
			name: "Success",
			setupMock: func(specialization *mock.MockSpecializationUsecase) {
				specialization.EXPECT().GetAllSpecializationNames(gomock.Any()).Return(validSpecializationNamesResponse(), nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Internal server error",
			setupMock: func(specialization *mock.MockSpecializationUsecase) {
				specialization.EXPECT().GetAllSpecializationNames(gomock.Any()).Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("database error")))
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

			mockSpecialization := mock.NewMockSpecializationUsecase(ctrl)
			tt.setupMock(mockSpecialization)

			handler := &SpecializationHandler{
				specialization: mockSpecialization,
			}

			req := httptest.NewRequest(http.MethodGet, "/specialization/all", nil)
			req = req.WithContext(req.Context())

			w := httptest.NewRecorder()

			handler.GetAllSpecializationNames(w, req)

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
				var response dto.SpecializationNamesResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, validSpecializationNamesResponse().Names, response.Names)
			}
		})
	}
}

func TestSpecializationHandler_GetSpecializationSalaries(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		mockSetup        func(specialization *mock.MockSpecializationUsecase)
		expectedStatus   int
		expectedResponse dto.SpecializationSalaryRangesResponse
		expectError      bool
		errorMessage     string
	}{
		{
			name: "успешное получение зарплат",
			mockSetup: func(specialization *mock.MockSpecializationUsecase) {
				resp := dto.SpecializationSalaryRangesResponse{
					Specializations: []dto.SpecializationSalaryRange{
						{ID: 1, Name: "Software Developer", MinSalary: 1000, MaxSalary: 5000, AvgSalary: 3000},
						{ID: 2, Name: "Data Scientist", MinSalary: 1500, MaxSalary: 6000, AvgSalary: 4000},
					},
				}
				specialization.EXPECT().
					GetSpecializationSalaries(gomock.Any()).
					Return(&resp, nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: dto.SpecializationSalaryRangesResponse{
				Specializations: []dto.SpecializationSalaryRange{
					{ID: 1, Name: "Software Developer", MinSalary: 1000, MaxSalary: 5000, AvgSalary: 3000},
					{ID: 2, Name: "Data Scientist", MinSalary: 1500, MaxSalary: 6000, AvgSalary: 4000},
				},
			},
		},
		{
			name: "ошибка при получении данных о зарплатах",
			mockSetup: func(specialization *mock.MockSpecializationUsecase) {
				specialization.EXPECT().
					GetSpecializationSalaries(gomock.Any()).
					Return(nil, fmt.Errorf("internal server error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
			errorMessage:   "internal server error",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpecialization := mock.NewMockSpecializationUsecase(ctrl)

			// Настройка мока
			tc.mockSetup(mockSpecialization)

			handler := NewSpecializationHandler(mockSpecialization)

			req := httptest.NewRequest(http.MethodGet, "/specializations/salaries", nil)
			w := httptest.NewRecorder()

			handler.GetSpecializationSalaries(w, req)

			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tc.expectedStatus, res.StatusCode)

			if !tc.expectError {
				var actual dto.SpecializationSalaryRangesResponse
				err := json.NewDecoder(res.Body).Decode(&actual)
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, actual)
			} else {
				var apiErr utils.APIError
				err := json.NewDecoder(res.Body).Decode(&apiErr)
				require.NoError(t, err)
				require.Equal(t, http.StatusInternalServerError, apiErr.Status)
				require.Equal(t, tc.errorMessage, apiErr.Message)
			}
		})
	}
}
