package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSpecializationService_GetAllSpecializationNames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		mockSetup      func(*mock.MockSpecializationRepository)
		expectedResult *dto.SpecializationNamesResponse
		expectedErr    error
	}{
		{
			name: "Успешное получение списка имен специализаций",
			mockSetup: func(sr *mock.MockSpecializationRepository) {
				sr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Specialization{
						{ID: 1, Name: "Backend разработка"},
						{ID: 2, Name: "Frontend разработка"},
						{ID: 3, Name: "DevOps"},
					}, nil)
			},
			expectedResult: &dto.SpecializationNamesResponse{
				Names: []string{"Backend разработка", "Frontend разработка", "DevOps"},
			},
			expectedErr: nil,
		},
		{
			name: "Пустой список специализаций",
			mockSetup: func(sr *mock.MockSpecializationRepository) {
				sr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Specialization{}, nil)
			},
			expectedResult: &dto.SpecializationNamesResponse{
				Names: []string{},
			},
			expectedErr: nil,
		},
		{
			name: "Ошибка получения списка специализаций",
			mockSetup: func(sr *mock.MockSpecializationRepository) {
				sr.EXPECT().
					GetAll(gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении списка специализаций"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка специализаций"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			tc.mockSetup(mockSpecRepo)

			service := NewSpecializationService(mockSpecRepo)
			ctx := context.Background()

			result, err := service.GetAllSpecializationNames(ctx)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var serviceErr entity.Error
				require.ErrorAs(t, err, &serviceErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
