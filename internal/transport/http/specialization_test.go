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
