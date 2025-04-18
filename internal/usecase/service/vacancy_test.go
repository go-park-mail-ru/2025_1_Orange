package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestVacanciesService_CreateVacancy(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		employerID     int
		request        *dto.VacancyCreate
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository)
		expectedResult *dto.VacancyResponse
		expectedErr    error
	}{
		{
			name:       "Успешное создание вакансии",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Backend Developer",
				Specialization: "Backend",
				WorkFormat:     "remote",
				Employment:     "full",
				SalaryFrom:     100000,
				SalaryTo:       200000,
				Skills:         []string{"Go", "SQL"},
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Поиск специализации
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend").
					Return(1, nil)

				// Создание вакансии
				vr.EXPECT().
					Create(gomock.Any(), &entity.Vacancy{
						Title:            "Backend Developer",
						IsActive:         true,
						EmployerID:       1,
						SpecializationID: 1,
						WorkFormat:       "remote",
						Employment:       "full",
						SalaryFrom:       100000,
						SalaryTo:         200000,
					}).
					Return(&entity.Vacancy{
						ID:               1,
						Title:            "Backend Developer",
						EmployerID:       1,
						SpecializationID: 1,
						WorkFormat:       "remote",
						Employment:       "full",
						SalaryFrom:       100000,
						SalaryTo:         200000,
						CreatedAt:        now,
						UpdatedAt:        now,
					}, nil)

				// Поиск навыков
				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "SQL"}).
					Return([]int{1, 2}, nil)

				// Добавление навыков
				vr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
					Return(nil)

				// Получение специализации
				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend",
					}, nil)

				// Получение навыков
				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
					}, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:             1,
				EmployerID:     1,
				Title:          "Backend Developer",
				Specialization: "Backend",
				WorkFormat:     "remote",
				Employment:     "full",
				SalaryFrom:     100000,
				SalaryTo:       200000,
				Skills:         []string{"Go", "SQL"},
				CreatedAt:      now.Format(time.RFC3339),
				UpdatedAt:      now.Format(time.RFC3339),
			},
			expectedErr: nil,
		},
		{
			name:       "Ошибка валидации вакансии",
			employerID: 1,
			request: &dto.VacancyCreate{
				SalaryFrom: 200000,
				SalaryTo:   100000, // Неправильный диапазон
			},
			mockSetup:      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository) {},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("минимальная зарплата не может быть больше максимальной"),
			),
		},
		{
			name:       "Ошибка при поиске специализации",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Backend Developer",
				Specialization: "Backend",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend").
					Return(0, fmt.Errorf("специализация не найдена"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("специализация не найдена"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := NewVacanciesService(mockVacancyRepo, mockApplicantRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.CreateVacancy(ctx, tc.employerID, tc.request)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestVacanciesService_GetVacancy(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		vacancyID      int
		currentUserID  int
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository)
		expectedResult *dto.VacancyResponse
		expectedErr    error
	}{
		{
			name:          "Успешное получение вакансии",
			vacancyID:     1,
			currentUserID: 2,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Получение вакансии
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:               1,
						Title:            "Backend Developer",
						EmployerID:       1,
						SpecializationID: 1,
						WorkFormat:       "remote",
						Employment:       "full",
						SalaryFrom:       100000,
						SalaryTo:         200000,
						CreatedAt:        now,
						UpdatedAt:        now,
					}, nil)

				// Получение специализации
				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend",
					}, nil)

				// Получение навыков
				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
					}, nil)

				// Проверка отклика
				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 2).
					Return(false, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:             1,
				Title:          "Backend Developer",
				EmployerID:     1,
				Specialization: "Backend",
				WorkFormat:     "remote",
				Employment:     "full",
				SalaryFrom:     100000,
				SalaryTo:       200000,
				Skills:         []string{"Go"},
				CreatedAt:      now.Format(time.RFC3339),
				UpdatedAt:      now.Format(time.RFC3339),
				Responded:      false,
			},
			expectedErr: nil,
		},
		{
			name:          "Вакансия не найдена",
			vacancyID:     999,
			currentUserID: 0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, fmt.Errorf("вакансия не найдена"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("вакансия не найдена"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := NewVacanciesService(mockVacancyRepo, mockApplicantRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.GetVacancy(ctx, tc.vacancyID, tc.currentUserID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestVacanciesService_UpdateVacancy(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		vacancyID      int
		request        *dto.VacancyUpdate
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository)
		expectedResult *dto.VacancyResponse
		expectedErr    error
	}{
		{
			name:      "Успешное обновление вакансии",
			vacancyID: 1,
			request: &dto.VacancyUpdate{
				Title:          "Updated Title",
				Specialization: "Backend",
				Skills:         []string{"Go", "PostgreSQL"},
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Получение существующей вакансии
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:         1,
						EmployerID: 1,
					}, nil)

				// Поиск специализации
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend").
					Return(1, nil)

				// Обновление вакансии
				vr.EXPECT().
					Update(gomock.Any(), &entity.Vacancy{
						ID:               1,
						EmployerID:       1,
						Title:            "Updated Title",
						SpecializationID: 1,
					}).
					Return(&entity.Vacancy{
						ID:               1,
						Title:            "Updated Title",
						EmployerID:       1,
						SpecializationID: 1,
						CreatedAt:        now,
						UpdatedAt:        now,
					}, nil)

				// Удаление старых навыков
				vr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				// Поиск новых навыков
				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "PostgreSQL"}).
					Return([]int{1, 3}, nil)

				// Добавление новых навыков
				vr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 3}).
					Return(nil)

				// Получение специализации
				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend",
					}, nil)

				// Получение навыков
				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 3, Name: "PostgreSQL"},
					}, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:             1,
				Title:          "Updated Title",
				EmployerID:     1,
				Specialization: "Backend",
				Skills:         []string{"Go", "PostgreSQL"},
				CreatedAt:      now.Format(time.RFC3339),
				UpdatedAt:      now.Format(time.RFC3339),
			},
			expectedErr: nil,
		},
		{
			name:      "Вакансия не найдена",
			vacancyID: 999,
			request:   &dto.VacancyUpdate{},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, fmt.Errorf("вакансия не найдена"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("вакансия не найдена"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := NewVacanciesService(mockVacancyRepo, mockApplicantRepo, mockSpecRepo)
			ctx := context.WithValue(context.Background(), "employerID", 1)

			result, err := service.UpdateVacancy(ctx, tc.vacancyID, tc.request)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestVacanciesService_DeleteVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		vacancyID      int
		employerID     int
		mockSetup      func(*mock.MockVacancyRepository)
		expectedResult *dto.DeleteVacancy
		expectedErr    error
	}{
		{
			name:       "Успешное удаление вакансии",
			vacancyID:  1,
			employerID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				// Получение вакансии
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:         1,
						EmployerID: 1,
					}, nil)

				// Удаление навыков
				vr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				// Удаление города
				vr.EXPECT().
					DeleteCity(gomock.Any(), 1).
					Return(nil)

				// Удаление вакансии
				vr.EXPECT().
					Delete(gomock.Any(), 1).
					Return(nil)
			},
			expectedResult: &dto.DeleteVacancy{
				Success: true,
				Message: "Вакансия с id=1 успешно удалена",
			},
			expectedErr: nil,
		},
		{
			name:       "Вакансия не принадлежит работодателю",
			vacancyID:  1,
			employerID: 2,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:         1,
						EmployerID: 1, // Другой работодатель
					}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrForbidden,
				fmt.Errorf("вакансия с id=1 не принадлежит работодателю с id=2"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(mockVacancyRepo, mockApplicantRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.DeleteVacancy(ctx, tc.vacancyID, tc.employerID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var serviceErr entity.Error
				require.ErrorAs(t, err, &serviceErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestVacanciesService_GetAll(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		currentUserID  int
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository)
		expectedResult []dto.VacancyShortResponse
		expectedErr    error
	}{
		{
			name:          "Успешное получение списка вакансий",
			currentUserID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Получение вакансий
				vr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Vacancy{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       1,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full",
							SalaryFrom:       100000,
							SalaryTo:         200000,
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				// Получение специализации
				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend",
					}, nil)

				// Проверка отклика
				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					EmployerID:     1,
					Specialization: "Backend",
					WorkFormat:     "remote",
					Employment:     "full",
					SalaryFrom:     100000,
					SalaryTo:       200000,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					Responded:      false,
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := NewVacanciesService(mockVacancyRepo, mockApplicantRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.GetAll(ctx, tc.currentUserID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestVacanciesService_ApplyToVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		vacancyID     int
		applicantID   int
		mockSetup     func(*mock.MockVacancyRepository)
		expectedError error
	}{
		{
			name:        "Успешный отклик на вакансию",
			vacancyID:   1,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				// Проверка существования вакансии
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1}, nil)

				// Проверка существования отклика
				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)

				// Создание отклика
				vr.EXPECT().
					CreateResponse(gomock.Any(), 1, 1).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "Вакансия не найдена",
			vacancyID:   999,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, fmt.Errorf("вакансия не найдена"))
			},
			expectedError: fmt.Errorf("vacancy not found: вакансия не найдена"),
		},
		{
			name:        "Повторный отклик",
			vacancyID:   1,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(true, nil)
			},
			expectedError: entity.NewError(
				entity.ErrAlreadyExists,
				fmt.Errorf("you have already applied to this vacancy"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(mockVacancyRepo, mockApplicantRepo, mockSpecRepo)
			ctx := context.Background()

			err := service.ApplyToVacancy(ctx, tc.vacancyID, tc.applicantID)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVacanciesService_GetActiveVacanciesByEmployerID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		employerID     int
		mockSetup      func(*mock.MockVacancyRepository)
		expectedResult []*dto.VacancyResponse
		expectedErr    error
	}{
		{
			name:       "Успешное получение активных вакансий",
			employerID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 1).
					Return([]*dto.VacancyResponse{
						{
							ID:         1,
							Title:      "Backend Developer",
							EmployerID: 1,
							CreatedAt:  now.Format(time.RFC3339),
						},
					}, nil)
			},
			expectedResult: []*dto.VacancyResponse{
				{
					ID:         1,
					Title:      "Backend Developer",
					EmployerID: 1,
					CreatedAt:  now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(mockVacancyRepo, mockApplicantRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.GetActiveVacanciesByEmployerID(ctx, tc.employerID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestVacanciesService_GetVacanciesByApplicantID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		applicantID    int
		mockSetup      func(*mock.MockVacancyRepository)
		expectedResult []*dto.VacancyResponse
		expectedErr    error
	}{
		{
			name:        "Успешное получение вакансий соискателя",
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 1).
					Return([]*dto.VacancyResponse{
						{
							ID:         1,
							Title:      "Backend Developer",
							EmployerID: 2,
							CreatedAt:  now.Format(time.RFC3339),
						},
					}, nil)
			},
			expectedResult: []*dto.VacancyResponse{
				{
					ID:         1,
					Title:      "Backend Developer",
					EmployerID: 2,
					CreatedAt:  now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(mockVacancyRepo, mockApplicantRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.GetVacanciesByApplicantID(ctx, tc.applicantID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
