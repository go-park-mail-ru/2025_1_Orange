//go:build ignore
// +build ignore

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
					Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, v *entity.Vacancy) (*entity.Vacancy, error) {
						require.Equal(t, "Backend Developer", v.Title)
						require.Equal(t, 1, v.SpecializationID)
						return &entity.Vacancy{
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
						}, nil
					})

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
		{
			name:       "Ошибка при создании вакансии",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Backend Developer",
				Specialization: "Backend",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend").
					Return(1, nil)

				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("ошибка создания вакансии"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка создания вакансии"),
		},
		{
			name:       "Ошибка при поиске навыков",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Backend Developer",
				Specialization: "Backend",
				Skills:         []string{"Go"},
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend").
					Return(1, nil)

				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go"}).
					Return(nil, fmt.Errorf("ошибка поиска навыков"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка поиска навыков"),
		},
		{
			name:       "Ошибка при добавлении навыков",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Backend Developer",
				Specialization: "Backend",
				Skills:         []string{"Go"},
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend").
					Return(1, nil)

				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1}).
					Return(fmt.Errorf("ошибка добавления навыков"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка добавления навыков"),
		},
		{
			name:       "Ошибка при получении специализации",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Backend Developer",
				Specialization: "Backend",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend").
					Return(1, nil)

				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{
						ID:               1,
						SpecializationID: 1,
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, fmt.Errorf("ошибка получения специализации"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка получения специализации"),
		},
		{
			name:       "Ошибка при получении навыков",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Backend Developer",
				Specialization: "Backend",
				Skills:         []string{"Go"},
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend").
					Return(1, nil)

				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1}).
					Return(nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{Name: "Backend"}, nil)

				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return(nil, fmt.Errorf("ошибка получения навыков"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка получения навыков"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := &VacanciesService{
				vacanciesRepository:      mockVacancyRepo,
				specializationRepository: mockSpecRepo,
			}

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
			name:          "Успешное получение списка вакансий (пользователь не авторизован)",
			currentUserID: 0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Получение списка вакансий
				vr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Vacancy{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       100,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full_time",
							SalaryFrom:       100000,
							SalaryTo:         150000,
							CreatedAt:        now,
							UpdatedAt:        now,
							City:             "Москва",
						},
						{
							ID:               2,
							Title:            "Frontend Developer",
							EmployerID:       101,
							SpecializationID: 2,
							WorkFormat:       "hybrid",
							Employment:       "part_time",
							SalaryFrom:       80000,
							SalaryTo:         120000,
							CreatedAt:        now.Add(-24 * time.Hour),
							UpdatedAt:        now.Add(-24 * time.Hour),
							City:             "Санкт-Петербург",
						},
					}, nil)

				// Получение специализации для первой вакансии
				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend",
					}, nil)

				// Получение специализации для второй вакансии
				sr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{
						ID:   2,
						Name: "Frontend",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					EmployerID:     100,
					Specialization: "Backend",
					WorkFormat:     "remote",
					Employment:     "full_time",
					SalaryFrom:     100000,
					SalaryTo:       150000,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Москва",
					Responded:      false,
				},
				{
					ID:             2,
					Title:          "Frontend Developer",
					EmployerID:     101,
					Specialization: "Frontend",
					WorkFormat:     "hybrid",
					Employment:     "part_time",
					SalaryFrom:     80000,
					SalaryTo:       120000,
					CreatedAt:      now.Add(-24 * time.Hour).Format(time.RFC3339),
					UpdatedAt:      now.Add(-24 * time.Hour).Format(time.RFC3339),
					City:           "Санкт-Петербург",
					Responded:      false,
				},
			},
			expectedErr: nil,
		},
		{
			name:          "Успешное получение списка вакансий (пользователь авторизован)",
			currentUserID: 200,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Получение списка вакансий
				vr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Vacancy{
						{
							ID: 1,
							// другие поля...
						},
					}, nil)

				// Получение специализации
				sr.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(&entity.Specialization{Name: "Backend"}, nil)

				// Проверка отклика пользователя
				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 200).
					Return(true, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:        1,
					Responded: true,
					// другие поля...
				},
			},
			expectedErr: nil,
		},
		{
			name:          "Ошибка при получении списка вакансий",
			currentUserID: 0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetAll(gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении списка вакансий"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка вакансий"),
			),
		},
		{
			name:          "Ошибка при получении специализации (пропускаем вакансию)",
			currentUserID: 0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Vacancy{
						{
							ID:               1,
							SpecializationID: 1,
							// другие поля...
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении специализации"),
					))
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:          "Ошибка при проверке отклика пользователя",
			currentUserID: 200,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Vacancy{
						{
							ID: 1,
							// другие поля...
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(&entity.Specialization{Name: "Backend"}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 200).
					Return(false, fmt.Errorf("ошибка проверки отклика"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка проверки отклика"),
		},
		{
			name:          "Вакансия без специализации",
			currentUserID: 0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Vacancy{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       100,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full_time",
							SalaryFrom:       100000,
							SalaryTo:         150000,
							CreatedAt:        now,
							UpdatedAt:        now,
							City:             "Москва",
						},
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID: 1,
					// другие поля...
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
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := &VacanciesService{
				vacanciesRepository:      mockVacancyRepo,
				specializationRepository: mockSpecRepo,
			}

			ctx := context.WithValue(context.Background(), "currentUserID", tc.currentUserID)
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
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository)
		expectedResult []dto.VacancyShortResponse
		expectedErr    error
	}{
		{
			name:       "Успешное получение активных вакансий работодателя",
			employerID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Мок получения активных вакансий
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 1).
					Return([]entity.Vacancy{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       1,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full_time",
							SalaryFrom:       100000,
							SalaryTo:         150000,
							TaxesIncluded:    true,
							WorkingHours:     40,
							CreatedAt:        now,
							UpdatedAt:        now,
							City:             "Москва",
						},
						{
							ID:               2,
							Title:            "Frontend Developer",
							EmployerID:       1,
							SpecializationID: 2,
							WorkFormat:       "hybrid",
							Employment:       "part_time",
							SalaryFrom:       80000,
							SalaryTo:         120000,
							CreatedAt:        now.Add(-time.Hour),
							UpdatedAt:        now.Add(-time.Hour),
							City:             "Санкт-Петербург",
						},
					}, nil)

				// Моки получения специализаций
				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend"}, nil)
				sr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{ID: 2, Name: "Frontend"}, nil)

				// Моки проверки откликов (для работодателя)
				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)
				vr.EXPECT().
					ResponseExists(gomock.Any(), 2, 1).
					Return(true, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					EmployerID:     1,
					Specialization: "Backend",
					WorkFormat:     "remote",
					Employment:     "full_time",
					SalaryFrom:     100000,
					SalaryTo:       150000,
					TaxesIncluded:  true,
					WorkingHours:   40,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Москва",
					Responded:      false,
				},
				{
					ID:             2,
					Title:          "Frontend Developer",
					EmployerID:     1,
					Specialization: "Frontend",
					WorkFormat:     "hybrid",
					Employment:     "part_time",
					SalaryFrom:     80000,
					SalaryTo:       120000,
					CreatedAt:      now.Add(-time.Hour).Format(time.RFC3339),
					UpdatedAt:      now.Add(-time.Hour).Format(time.RFC3339),
					City:           "Санкт-Петербург",
					Responded:      true,
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Нет активных вакансий у работодателя",
			employerID: 2,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 2).
					Return([]entity.Vacancy{}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:       "Ошибка при получении вакансий",
			employerID: 3,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 3).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении вакансий"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении вакансий"),
			),
		},
		{
			name:       "Ошибка при получении специализации (пропускаем вакансию)",
			employerID: 4,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 4).
					Return([]entity.Vacancy{
						{
							ID:               1,
							SpecializationID: 1,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении специализации"),
					))

				// ResponseExists не должен вызываться из-за continue
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:       "Ошибка при проверке отклика",
			employerID: 5,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 5).
					Return([]entity.Vacancy{
						{
							ID: 1,
						},
					}, nil)

				// Для вакансии без специализации GetByID не вызывается

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 5).
					Return(false, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при проверке отклика"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке отклика"),
			),
		},
		{
			name:       "Вакансия без специализации",
			employerID: 6,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 6).
					Return([]entity.Vacancy{
						{
							ID:         1,
							Title:      "General Position",
							EmployerID: 6,
						},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 6).
					Return(false, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:         1,
					Title:      "General Position",
					EmployerID: 6,
					Responded:  false,
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
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := &VacanciesService{
				vacanciesRepository:      mockVacancyRepo,
				specializationRepository: mockSpecRepo,
			}

			ctx := context.Background()
			result, err := service.GetActiveVacanciesByEmployerID(ctx, tc.employerID)

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

func TestVacanciesService_GetVacanciesByApplicantID_ResponseFormation(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		applicantID    int
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository)
		expectedResult []dto.VacancyShortResponse
		expectedErr    error
	}{
		{
			name:        "Вакансия со специализацией и откликом",
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 1).
					Return([]entity.Vacancy{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       100,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full_time",
							SalaryFrom:       100000,
							SalaryTo:         150000,
							TaxesIncluded:    true,
							WorkingHours:     40,
							CreatedAt:        now,
							UpdatedAt:        now,
							City:             "Москва",
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{Name: "Backend"}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(true, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					EmployerID:     100,
					Specialization: "Backend",
					WorkFormat:     "remote",
					Employment:     "full_time",
					SalaryFrom:     100000,
					SalaryTo:       150000,
					TaxesIncluded:  true,
					WorkingHours:   40,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Москва",
					Responded:      true,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Вакансия без специализации",
			applicantID: 2,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 2).
					Return([]entity.Vacancy{
						{
							ID:         2,
							Title:      "No Specialization",
							EmployerID: 101,
						},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 2, 2).
					Return(false, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:         2,
					Title:      "No Specialization",
					EmployerID: 101,
					Responded:  false,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка получения специализации (пропуск вакансии)",
			applicantID: 3,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 3).
					Return([]entity.Vacancy{
						{
							ID:               3,
							SpecializationID: 2,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(nil, fmt.Errorf("специализация не найдена"))
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:        "Неавторизованный пользователь (responded=false)",
			applicantID: 0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 0).
					Return([]entity.Vacancy{
						{
							ID: 4,
						},
					}, nil)

				// Не должно быть вызова ResponseExists для applicantID=0
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:        4,
					Responded: false,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Все поля вакансии заполнены",
			applicantID: 5,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 5).
					Return([]entity.Vacancy{
						{
							ID:               5,
							Title:            "Full Fields",
							EmployerID:       102,
							SpecializationID: 3,
							WorkFormat:       "office",
							Employment:       "part_time",
							SalaryFrom:       50000,
							SalaryTo:         80000,
							TaxesIncluded:    false,
							WorkingHours:     20,
							CreatedAt:        now,
							UpdatedAt:        now,
							City:             "Казань",
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 3).
					Return(&entity.Specialization{Name: "Fullstack"}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 5, 5).
					Return(false, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             5,
					Title:          "Full Fields",
					EmployerID:     102,
					Specialization: "Fullstack",
					WorkFormat:     "office",
					Employment:     "part_time",
					SalaryFrom:     50000,
					SalaryTo:       80000,
					TaxesIncluded:  false,
					WorkingHours:   20,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Казань",
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
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := &VacanciesService{
				vacanciesRepository:      mockVacancyRepo,
				specializationRepository: mockSpecRepo,
			}

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
