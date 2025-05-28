package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	m "ResuMatch/internal/usecase/mock"
	"context"
	"errors"
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
			name:       "Успешное создание вакансии со всеми полями",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:                "Backend Developer",
				Specialization:       "Backend разработка",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Описание вакансии",
				Tasks:                "Задачи вакансии",
				Requirements:         "Требования",
				OptionalRequirements: "Опциональные требования для кандидата",
				Skills:               []string{"Go", "PostgreSQL"},
				City:                 "Москва",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Мок для поиска специализации
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				// Мок для создания вакансии
				vr.EXPECT().
					Create(gomock.Any(), &entity.Vacancy{
						Title:                "Backend Developer",
						IsActive:             true,
						EmployerID:           1,
						SpecializationID:     1,
						WorkFormat:           "remote",
						Employment:           "full_time",
						Schedule:             "5/2",
						WorkingHours:         18,
						SalaryFrom:           100000,
						SalaryTo:             200000,
						TaxesIncluded:        true,
						Experience:           "3_6_years",
						Description:          "Описание вакансии",
						Tasks:                "Задачи вакансии",
						Requirements:         "Требования",
						OptionalRequirements: "Опциональные требования для кандидата",
						City:                 "Москва",
					}).
					Return(&entity.Vacancy{
						ID:                   1,
						EmployerID:           1,
						Title:                "Backend Developer",
						IsActive:             true,
						SpecializationID:     1,
						WorkFormat:           "remote",
						Employment:           "full_time",
						Schedule:             "5/2",
						WorkingHours:         18,
						SalaryFrom:           100000,
						SalaryTo:             200000,
						TaxesIncluded:        true,
						Experience:           "3_6_years",
						Description:          "Описание вакансии",
						Tasks:                "Задачи вакансии",
						Requirements:         "Требования",
						OptionalRequirements: "Опциональные требования для кандидата",
						City:                 "Москва",
						CreatedAt:            now,
						UpdatedAt:            now,
					}, nil)

				// Мок для поиска ID навыков
				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "PostgreSQL"}).
					Return([]int{1, 2}, nil)

				// Мок для добавления навыков к вакансии
				vr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
					Return(nil)

				// Мок для получения названия специализации
				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				// Мок для получения навыков вакансии
				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "PostgreSQL"},
					}, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:                   1,
				EmployerID:           1,
				Title:                "Backend Developer",
				Specialization:       "Backend разработка",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				City:                 "Москва",
				Description:          "Описание вакансии",
				Tasks:                "Задачи вакансии",
				Requirements:         "Требования",
				Skills:               []string{"Go", "PostgreSQL"},
				OptionalRequirements: "Опциональные требования для кандидата",
				CreatedAt:            now.Format(time.RFC3339),
				UpdatedAt:            now.Format(time.RFC3339),
			},
			expectedErr: nil,
		},
		{
			name:       "Ошибка поиска специализации",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Developer",
				Specialization: "Backend разработка",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(0, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при поиске специализации"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске специализации"),
			),
		},
		{
			name:       "Ошибка создания вакансии",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:                "Backend Developer",
				Specialization:       "Backend разработка",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Описание вакансии",
				Tasks:                "Задачи вакансии",
				Requirements:         "Требования",
				OptionalRequirements: "Опциональные требования для кандидата",
				Skills:               []string{"Go", "PostgreSQL"},
				City:                 "Москва",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при создании вакансии"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании вакансии"),
			),
		},
		{
			name:       "Ошибка добавления навыков",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:                "Backend Developer",
				Specialization:       "Backend разработка",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "3_6_years",
				Description:          "Описание вакансии",
				Tasks:                "Задачи вакансии",
				Requirements:         "Требования",
				OptionalRequirements: "Aaaaaaaffffffffff",
				Skills:               []string{"Go", "PostgreSQL"},
				City:                 "Москва",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "PostgreSQL"}).
					Return([]int{1, 2}, nil)

				vr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
					Return(entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("описание дополнительных требований должно быть от 10 до 500 символов"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("описание дополнительных требований должно быть от 10 до 500 символов"),
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
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
				mockResumeRepo,
				mockApplicantService,
			)
			ctx := context.Background()

			result, err := service.CreateVacancy(ctx, tc.employerID, tc.request)

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
func TestVacanciesService_GetVacancy(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		id             int
		currentUserID  int
		userRole       string
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult *dto.VacancyResponse
		expectedErr    error
	}{
		{
			name:          "Успешное получение вакансии для соискателя",
			id:            1,
			currentUserID: 2,
			userRole:      "applicant",
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:                   1,
						EmployerID:           1,
						Title:                "Backend Developer",
						SpecializationID:     1,
						WorkFormat:           "remote",
						Employment:           "full",
						Schedule:             "flexible",
						WorkingHours:         18,
						SalaryFrom:           100000,
						SalaryTo:             200000,
						TaxesIncluded:        true,
						Experience:           "3-5 years",
						Description:          "Описание вакансии",
						Tasks:                "Задачи вакансии",
						Requirements:         "Требования вакансии",
						OptionalRequirements: "Дополнительные требования",
						CreatedAt:            now,
						UpdatedAt:            now,
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 2).
					Return(false, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 2).
					Return(true, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:                   1,
				EmployerID:           1,
				Title:                "Backend Developer",
				Specialization:       "Backend разработка",
				WorkFormat:           "remote",
				Employment:           "full",
				Schedule:             "flexible",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "3-5 years",
				Description:          "Описание вакансии",
				Tasks:                "Задачи вакансии",
				Requirements:         "Требования вакансии",
				OptionalRequirements: "Дополнительные требования",
				Skills:               []string{"Go", "SQL"},
				CreatedAt:            now.Format(time.RFC3339),
				UpdatedAt:            now.Format(time.RFC3339),
				Responded:            false,
				Liked:                true,
			},
			expectedErr: nil,
		},
		{
			name:          "Успешное получение вакансии для соискателя",
			id:            1,
			currentUserID: 2,
			userRole:      "applicant",
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:                   1,
						EmployerID:           1,
						Title:                "Backend Developer",
						SpecializationID:     1,
						WorkFormat:           "remote",
						Employment:           "full",
						Schedule:             "flexible",
						WorkingHours:         18,
						SalaryFrom:           100000,
						SalaryTo:             200000,
						TaxesIncluded:        true,
						Experience:           "3-5 years",
						Description:          "Описание вакансии",
						Tasks:                "Задачи вакансии",
						Requirements:         "Требования вакансии",
						OptionalRequirements: "Дополнительные требования",
						CreatedAt:            now,
						UpdatedAt:            now,
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 2).
					Return(false, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 2).
					Return(true, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:                   1,
				EmployerID:           1,
				Title:                "Backend Developer",
				Specialization:       "Backend разработка",
				WorkFormat:           "remote",
				Employment:           "full",
				Schedule:             "flexible",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "3-5 years",
				Description:          "Описание вакансии",
				Tasks:                "Задачи вакансии",
				Requirements:         "Требования вакансии",
				OptionalRequirements: "Дополнительные требования",
				Skills:               []string{"Go", "SQL"},
				CreatedAt:            now.Format(time.RFC3339),
				UpdatedAt:            now.Format(time.RFC3339),
				Responded:            false,
				Liked:                true,
			},
			expectedErr: nil,
		},
		{
			name:          "Вакансия не найдена",
			id:            999,
			currentUserID: 0,
			userRole:      "",
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(&entity.Vacancy{}, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("вакансия с id=999 не найдена"),
					))

			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("вакансия с id=999 не найдена"),
			),
		},
		{
			name:          "Ошибка при получении специализации",
			id:            1,
			currentUserID: 0,
			userRole:      "",
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						SpecializationID: 1,
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении специализации"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении специализации"),
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
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
				mockResumeRepo,
				mockApplicantService,
			)
			ctx := context.Background()

			result, err := service.GetVacancy(ctx, tc.id, tc.currentUserID, tc.userRole)

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

func TestVacanciesService_UpdateVacancy(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		id             int
		employerID     int
		request        *dto.VacancyUpdate
		mockSetup      func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository)
		expectedResult *dto.VacancyResponse
		expectedErr    error
	}{
		{
			name:       "Успешное обновление вакансии с навыками и специализацией",
			id:         1,
			employerID: 10,
			request: &dto.VacancyUpdate{
				Title:                "Updated Title",
				Specialization:       "IT",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "1_3_years",
				Description:          "Updated description",
				Tasks:                "Updated tasks",
				Requirements:         "Updated requirements",
				OptionalRequirements: "Optional stuff",
				Skills:               []string{"Go", "Docker"},
				City:                 "Moscow",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:         1,
						EmployerID: 10,
					}, nil)

				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "IT").
					Return(42, nil)

				vr.EXPECT().
					Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
					DoAndReturn(func(_ context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
						vacancy.CreatedAt = now
						vacancy.UpdatedAt = now
						return vacancy, nil
					})

				vr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "Docker"}).
					Return([]int{1, 2}, nil)

				vr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
					Return(nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 42).
					Return(&entity.Specialization{Name: "IT"}, nil)

				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return([]entity.Skill{
						{Name: "Go"},
						{Name: "Docker"},
					}, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:                   1,
				EmployerID:           10,
				Title:                "Updated Title",
				Specialization:       "IT",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "1_3_years",
				Description:          "Updated description",
				Tasks:                "Updated tasks",
				Requirements:         "Updated requirements",
				OptionalRequirements: "Optional stuff",
				CreatedAt:            now.Format(time.RFC3339),
				UpdatedAt:            now.Format(time.RFC3339),
				Skills:               []string{"Go", "Docker"},
				City:                 "Moscow",
			},
			expectedErr: nil,
		},
		{
			name:       "Ошибка: вакансия не найдена",
			id:         1,
			employerID: 10,
			request: &dto.VacancyUpdate{
				Title:                "Updated Title",
				Specialization:       "IT",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "1_3_years",
				Description:          "Updated description",
				Tasks:                "Updated tasks",
				Requirements:         "Updated requirements",
				OptionalRequirements: "Optional stuff",
				Skills:               []string{"Go", "Docker"},
				City:                 "Moscow",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, fmt.Errorf("not found"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("not found"),
		},
		{
			name:       "Ошибка: вакансия принадлежит другому работодателю",
			id:         1,
			employerID: 10,
			request: &dto.VacancyUpdate{
				Title:                "Updated Title",
				Specialization:       "IT",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "1_3_years",
				Description:          "Updated description",
				Tasks:                "Updated tasks",
				Requirements:         "Updated requirements",
				OptionalRequirements: "Optional stuff",
				Skills:               []string{"Go", "Docker"},
				City:                 "Moscow",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1, EmployerID: 99}, nil)
			},
			expectedResult: nil,
			expectedErr:    entity.NewError(entity.ErrForbidden, fmt.Errorf("вакансия с id=1 не принадлежит работодателю с id=10")),
		},
		{
			name:       "Ошибка при поиске специализации",
			id:         1,
			employerID: 10,
			request: &dto.VacancyUpdate{
				Title:                "Updated Title",
				Specialization:       "IT",
				WorkFormat:           "remote",
				Employment:           "full_time",
				Schedule:             "5/2",
				WorkingHours:         18,
				SalaryFrom:           100000,
				SalaryTo:             200000,
				TaxesIncluded:        true,
				Experience:           "1_3_years",
				Description:          "Updated description",
				Tasks:                "Updated tasks",
				Requirements:         "Updated requirements",
				OptionalRequirements: "Optional stuff",
				Skills:               []string{"Go", "Docker"},
				City:                 "Moscow",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)

				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "IT").
					Return(0, fmt.Errorf("not found"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("not found"),
		},
		// {
		// 	name:       "Ошибка при удалении навыков",
		// 	id:         1,
		// 	employerID: 10,
		// 	request: &dto.VacancyUpdate{
		// 		Title:                "Updated Title",
		// 		Specialization:       "IT",
		// 		WorkFormat:           "remote",
		// 		Employment:           "full_time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         18,
		// 		SalaryFrom:           100000,
		// 		SalaryTo:             200000,
		// 		TaxesIncluded:        true,
		// 		Experience:           "1_3_years",
		// 		Description:          "Updated description",
		// 		Tasks:                "Updated tasks",
		// 		Requirements:         "Updated requirements",
		// 		OptionalRequirements: "Optional stuff",
		// 		Skills:               []string{"Go", "Docker"},
		// 		City:                 "Moscow",
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)

		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "").
		// 			Return(0, nil)

		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10, CreatedAt: now, UpdatedAt: now}, nil)

		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 1).
		// 			Return(fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"),
		// },

		// {
		// 	name:       "Ошибка: вакансия принадлежит другому работодателю",
		// 	id:         1,
		// 	employerID: 10,
		// 	request:    &dto.VacancyUpdate{},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 99}, nil)
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    entity.NewError(entity.ErrForbidden, fmt.Errorf("вакансия с id=1 не принадлежит работодателю с id=10")),
		// },
		// {
		// 	name:       "Ошибка при поиске специализации",
		// 	id:         1,
		// 	employerID: 10,
		// 	request: &dto.VacancyUpdate{
		// 		Specialization: "NonExist",
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)

		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "NonExist").
		// 			Return(0, fmt.Errorf("not found"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("not found"),
		// },
		// {
		// 	name:       "Ошибка при удалении навыков",
		// 	id:         1,
		// 	employerID: 10,
		// 	request: &dto.VacancyUpdate{
		// 		Skills: []string{"Go"},
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)

		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "").
		// 			Return(0, nil)

		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10, CreatedAt: now, UpdatedAt: now}, nil)

		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 1).
		// 			Return(fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"),
		// },
		// {
		// 	name:       "Ошибка при поиске ID навыков",
		// 	id:         1,
		// 	employerID: 10,
		// 	request: &dto.VacancyUpdate{
		// 		Title:                "Aaaaaaaaaaaa",
		// 		Specialization:       "Aaaaaaaaaaaa",
		// 		WorkFormat:           "office",
		// 		Employment:           "full_time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         1,
		// 		SalaryFrom:           100000,
		// 		SalaryTo:             150000,
		// 		TaxesIncluded:        true,
		// 		Experience:           "1_3_years",
		// 		City:                 "Aaaaaaaaaaaa",
		// 		Skills:               []string{"Aaaaaaaaaaaa"},
		// 		Description:          "Aaaaaaaaaaaa",
		// 		Tasks:                "Aaaaaaaaaaaa",
		// 		Requirements:         "Aaaaaaaaaaaa",
		// 		OptionalRequirements: "Aaaaaaaaaaaa",
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)

		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "Aaaaaaaaaaaa").
		// 			Return(1, nil)

		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10, CreatedAt: now, UpdatedAt: now}, nil)

		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 1).
		// 			Return(nil)

		// 		vr.EXPECT().
		// 			FindSkillIDsByNames(gomock.Any(), []string{"Aaaaaaaaaaaa"}).
		// 			Return(nil, fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"),
		// },
		// {
		// 	name:       "Ошибка при добавлении навыков",
		// 	id:         1,
		// 	employerID: 10,
		// 	request: &dto.VacancyUpdate{
		// 		Skills: []string{"Go"},
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)

		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "").
		// 			Return(0, nil)

		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10, CreatedAt: now, UpdatedAt: now}, nil)

		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 1).
		// 			Return(nil)

		// 		vr.EXPECT().
		// 			FindSkillIDsByNames(gomock.Any(), []string{"Go"}).
		// 			Return([]int{1}, nil)

		// 		vr.EXPECT().
		// 			AddSkills(gomock.Any(), 1, []int{1}).
		// 			Return(fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"),
		// },
		// {
		// 	name:       "Ошибка при получении специализации после обновления",
		// 	id:         1,
		// 	employerID: 10,
		// 	request: &dto.VacancyUpdate{
		// 		Title:                "Updated Title",
		// 		Specialization:       "Finance",
		// 		WorkFormat:           "remote",
		// 		Employment:           "full_time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         18,
		// 		SalaryFrom:           100000,
		// 		SalaryTo:             200000,
		// 		TaxesIncluded:        true,
		// 		Experience:           "1_3_years",
		// 		Description:          "Updated description",
		// 		Tasks:                "Updated tasks",
		// 		Requirements:         "Updated requirements",
		// 		OptionalRequirements: "Optional stuff",
		// 		Skills:               []string{"Go", "Docker"},
		// 		City:                 "Moscow",
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)

		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "Finance").
		// 			Return(42, nil)

		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10, SpecializationID: 42, CreatedAt: now, UpdatedAt: now}, nil)

		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 1).
		// 			Return(nil)

		// 		vr.EXPECT().
		// 			GetSkillsByVacancyID(gomock.Any(), 1).
		// 			Return([]entity.Skill{}, nil)

		// 		sr.EXPECT().
		// 			GetByID(gomock.Any(), 42).
		// 			Return(nil, fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"),
		// },
		// {
		// 	name:       "Ошибка валидации вакансии",
		// 	id:         1,
		// 	employerID: 10,
		// 	request: &dto.VacancyUpdate{
		// 		Title:                "O",
		// 		Specialization:       "IT",
		// 		WorkFormat:           "remote",
		// 		Employment:           "full_time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         18,
		// 		SalaryFrom:           100000,
		// 		SalaryTo:             200000,
		// 		TaxesIncluded:        true,
		// 		Experience:           "1_3_years",
		// 		Description:          "Updated description",
		// 		Tasks:                "Updated tasks",
		// 		Requirements:         "Updated requirements",
		// 		OptionalRequirements: "Optional stuff",
		// 		Skills:               []string{"Go", "Docker"},
		// 		City:                 "Moscow",
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)

		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "IT").
		// 			Return(0, nil)

		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			DoAndReturn(func(_ context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
		// 				return nil, fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов")
		// 			})
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"),
		// },
		// {
		// 	name:       "Успешное обновление вакансии с навыками и специализацией",
		// 	id:         1,
		// 	employerID: 10,
		// 	request: &dto.VacancyUpdate{
		// 		Title:                "Updated Title",
		// 		Specialization:       "IT",
		// 		WorkFormat:           "remote",
		// 		Employment:           "full_time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         18,
		// 		SalaryFrom:           100000,
		// 		SalaryTo:             200000,
		// 		TaxesIncluded:        true,
		// 		Experience:           "1_3_years",
		// 		Description:          "Updated description",
		// 		Tasks:                "Updated tasks",
		// 		Requirements:         "Updated requirements",
		// 		OptionalRequirements: "Optional stuff",
		// 		City:                 "Moscow",
		// 		Skills:               []string{"Go", "Docker"},
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		// Получаем существующую вакансию.
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 1).
		// 			Return(&entity.Vacancy{ID: 1, EmployerID: 10}, nil)
		// 		// Если специализация задана, получаем её ID.
		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "IT").
		// 			Return(42, nil)
		// 		// Валидируем и обновляем вакансию.
		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			DoAndReturn(func(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
		// 				vacancy.CreatedAt = now
		// 				vacancy.UpdatedAt = now
		// 				return vacancy, nil
		// 			})
		// 		// Удаляем старые навыки.
		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 1).
		// 			Return(nil)
		// 		// Определяем ID навыков по именам.
		// 		vr.EXPECT().
		// 			FindSkillIDsByNames(gomock.Any(), []string{"Go", "Docker"}).
		// 			Return([]int{1, 2}, nil)
		// 		// Добавляем навыки.
		// 		vr.EXPECT().
		// 			AddSkills(gomock.Any(), 1, []int{1, 2}).
		// 			Return(nil)
		// 		// Получаем данные специализации.
		// 		sr.EXPECT().
		// 			GetByID(gomock.Any(), 42).
		// 			Return(&entity.Specialization{Name: "IT"}, nil)
		// 		// Получаем навыки вакансии.
		// 		vr.EXPECT().
		// 			GetSkillsByVacancyID(gomock.Any(), 1).
		// 			Return([]entity.Skill{{Name: "Go"}, {Name: "Docker"}}, nil)
		// 	},
		// 	expectedResult: &dto.VacancyResponse{
		// 		ID:                   1,
		// 		EmployerID:           10,
		// 		Title:                "Updated Title",
		// 		Specialization:       "IT",
		// 		WorkFormat:           "remote",
		// 		Employment:           "full_time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         18,
		// 		SalaryFrom:           100000,
		// 		SalaryTo:             200000,
		// 		TaxesIncluded:        true,
		// 		Experience:           "1_3_years",
		// 		Description:          "Updated description",
		// 		Tasks:                "Updated tasks",
		// 		Requirements:         "Updated requirements",
		// 		OptionalRequirements: "Optional stuff",
		// 		City:                 "Moscow",
		// 		CreatedAt:            now.Format(time.RFC3339),
		// 		UpdatedAt:            now.Format(time.RFC3339),
		// 		Skills:               []string{"Go", "Docker"},
		// 	},
		// 	expectedErr: nil,
		// },
		// {
		// 	name:       "Ошибка: вакансия не найдена",
		// 	id:         3,
		// 	employerID: 30,
		// 	request:    &dto.VacancyUpdate{},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 3).
		// 			Return(nil, fmt.Errorf("vacancy not found"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("vacancy not found"),
		// },
		// {
		// 	name:       "Ошибка: вакансия принадлежит другому работодателю",
		// 	id:         4,
		// 	employerID: 40,
		// 	request:    &dto.VacancyUpdate{Title: "Some Title"},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 4).
		// 			Return(&entity.Vacancy{ID: 4, EmployerID: 999}, nil)
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr: entity.NewError(entity.ErrForbidden,
		// 		fmt.Errorf("вакансия с id=4 не принадлежит работодателю с id=40")),
		// },
		// {
		// 	name:       "Ошибка при поиске специализации",
		// 	id:         5,
		// 	employerID: 50,
		// 	request:    &dto.VacancyUpdate{Title: "Title", Specialization: "NonExist"},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 5).
		// 			Return(&entity.Vacancy{ID: 5, EmployerID: 50}, nil)
		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "NonExist").
		// 			Return(0, fmt.Errorf("specialization not found"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("specialization not found"),
		// },
		// {
		// 	name:       "Ошибка валидации вакансии",
		// 	id:         6,
		// 	employerID: 60,
		// 	request:    &dto.VacancyUpdate{Title: "", Specialization: ""},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 6).
		// 			Return(&entity.Vacancy{ID: 6, EmployerID: 60}, nil)
		// 		// Здесь не вызывается Update, так как vacancy.Validate() (предположительно) вернёт ошибку для пустого Title.
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nназвание вакансии должно быть от 3 до 50 символов"),
		// },
		// {
		// 	name:       "Ошибка при обновлении вакансии",
		// 	id:         7,
		// 	employerID: 70,
		// 	request: &dto.VacancyUpdate{
		// 		Title:                "Title",
		// 		Specialization:       "IT",
		// 		WorkFormat:           "remote",
		// 		Employment:           "full-time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         10,
		// 		SalaryFrom:           1500,
		// 		SalaryTo:             2500,
		// 		TaxesIncluded:        true,
		// 		Experience:           "6_plus_years",
		// 		Description:          "Desc",
		// 		Tasks:                "Tasks",
		// 		Requirements:         "Req",
		// 		OptionalRequirements: "Optional",
		// 		City:                 "City",
		// 		Skills:               []string{},
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 7).
		// 			Return(&entity.Vacancy{ID: 7, EmployerID: 70}, nil)
		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			Return(nil, fmt.Errorf("bad request\nнекорректный ID специализации"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nнекорректный ID специализации"),
		// },

		// {
		// 	name:       "Ошибка при поиске SkillIDs",
		// 	id:         9,
		// 	employerID: 90,
		// 	request: &dto.VacancyUpdate{
		// 		Title:                "Updated Title",
		// 		Specialization:       "IT",
		// 		WorkFormat:           "remote",
		// 		Employment:           "full_time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         18,
		// 		SalaryFrom:           100000,
		// 		SalaryTo:             200000,
		// 		TaxesIncluded:        true,
		// 		Experience:           "1_3_years",
		// 		Description:          "Updated description",
		// 		Tasks:                "Updated tasks",
		// 		Requirements:         "Updated requirements",
		// 		OptionalRequirements: "Optional stuff",
		// 		City:                 "Moscow",
		// 		Skills:               []string{"Go", "Docker"},
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 9).
		// 			Return(&entity.Vacancy{ID: 9, EmployerID: 90}, nil)
		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			DoAndReturn(func(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
		// 				vacancy.CreatedAt = now
		// 				vacancy.UpdatedAt = now
		// 				return vacancy, nil
		// 			})
		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 9).
		// 			Times(0)
		// 		vr.EXPECT().
		// 			FindSkillIDsByNames(gomock.Any(), []string{"Go", "Docker"}).
		// 			Times(0)
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nнекорректный ID специализации"),
		// },
		// {
		// 	name:       "Ошибка при добавлении навыков",
		// 	id:         10,
		// 	employerID: 100,
		// 	request: &dto.VacancyUpdate{
		// 		Title:          "Title",
		// 		Specialization: "",
		// 		Skills:         []string{"Kubernetes"},
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 10).
		// 			Return(&entity.Vacancy{ID: 10, EmployerID: 100}, nil)
		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			DoAndReturn(func(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
		// 				vacancy.CreatedAt = now
		// 				vacancy.UpdatedAt = now
		// 				return vacancy, nil
		// 			})
		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 10).
		// 			Return(nil)
		// 		vr.EXPECT().
		// 			FindSkillIDsByNames(gomock.Any(), []string{"Kubernetes"}).
		// 			Return([]int{3}, nil)
		// 		vr.EXPECT().
		// 			AddSkills(gomock.Any(), 10, []int{1}).
		// 			Return(fmt.Errorf("bad request\nнекорректный ID специализации"))
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nнекорректный ID специализации"),
		// },
		// {
		// 	name:       "Ошибка получения специализации",
		// 	id:         11,
		// 	employerID: 110,
		// 	request: &dto.VacancyUpdate{
		// 		Title:                "Updated Title",
		// 		Specialization:       "Finance",
		// 		WorkFormat:           "remote",
		// 		Employment:           "full_time",
		// 		Schedule:             "5/2",
		// 		WorkingHours:         18,
		// 		SalaryFrom:           100000,
		// 		SalaryTo:             200000,
		// 		TaxesIncluded:        true,
		// 		Experience:           "1_3_years",
		// 		Description:          "Updated description",
		// 		Tasks:                "Updated tasks",
		// 		Requirements:         "Updated requirements",
		// 		OptionalRequirements: "Optional stuff",
		// 		Skills:               []string{"Go", "Docker"},
		// 		City:                 "Moscow",
		// 	},
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
		// 		vr.EXPECT().
		// 			GetByID(gomock.Any(), 11).
		// 			Return(&entity.Vacancy{ID: 11, EmployerID: 110}, nil)
		// 		vr.EXPECT().
		// 			FindSpecializationIDByName(gomock.Any(), "Finance").
		// 			Return(55, nil)
		// 		vr.EXPECT().
		// 			Update(gomock.Any(), gomock.AssignableToTypeOf(&entity.Vacancy{})).
		// 			DoAndReturn(func(ctx context.Context, vacancy *entity.Vacancy) (*entity.Vacancy, error) {
		// 				vacancy.CreatedAt = now
		// 				vacancy.UpdatedAt = now

		// 				vacancy.SpecializationID = 55
		// 				return vacancy, nil
		// 			})
		// 		vr.EXPECT().
		// 			DeleteSkills(gomock.Any(), 11).
		// 			Times(0)

		// 		sr.EXPECT().
		// 			GetByID(gomock.Any(), 55).
		// 			Times(0)
		// 	},
		// 	expectedResult: nil,
		// 	expectedErr:    fmt.Errorf("bad request\nнекорректный формат работы"),
		// },
		// Можно добавить другие кейсы, например ошибки FindSkillIDsByNames, AddSkills, GetByID specializationRepository, GetSkillsByVacancyID и т.д.
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

			service := NewVacanciesService(
				mockVacancyRepo,
				nil, // applicantRepo
				mockSpecRepo,
				nil, // employerService
				nil, // resumeRepo
				nil, // applicantService
			)

			ctx := context.Background()
			resp, err := service.UpdateVacancy(ctx, tc.id, tc.employerID, tc.request)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.Equal(t, tc.expectedResult, resp)
			}
		})
	}
}

// func TestVacanciesService_SearchVacanciesByQueryAndSpecializations(t *testing.T) {
// 	t.Parallel()

// 	now := time.Now()

// 	testCases := []struct {
// 		name            string
// 		userID          int
// 		userRole        string
// 		searchQuery     string
// 		specializations []string
// 		limit           int
// 		offset          int
// 		mockSetup       func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
// 		expectedResult  []dto.VacancyShortResponse
// 		expectedErr     error
// 	}{
// 		{
// 			name:            "Успешный поиск по запросу и специализациям для соискателя",
// 			userID:          1,
// 			userRole:        "applicant",
// 			searchQuery:     "backend",
// 			specializations: []string{"Backend разработка"},
// 			limit:           10,
// 			offset:          0,
// 			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
// 				vr.EXPECT().
// 					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
// 					Return([]int{1}, nil)

// 				vr.EXPECT().
// 					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "backend", []int{1}, 10, 0).
// 					Return([]*entity.Vacancy{
// 						{
// 							ID:               1,
// 							Title:            "Senior Backend Developer",
// 							EmployerID:       1,
// 							SpecializationID: 1,
// 							WorkFormat:       "remote",
// 							Employment:       "full",
// 							WorkingHours:     18,
// 							SalaryFrom:       150000,
// 							SalaryTo:         250000,
// 							TaxesIncluded:    true,
// 							City:             "Москва",
// 							CreatedAt:        now,
// 							UpdatedAt:        now,
// 						},
// 					}, nil)

// 				sr.EXPECT().
// 					GetByID(gomock.Any(), 1).
// 					Return(&entity.Specialization{
// 						ID:   1,
// 						Name: "Backend разработка",
// 					}, nil)

// 				vr.EXPECT().
// 					ResponseExists(gomock.Any(), 1, 1).
// 					Return(false, nil)

// 				vr.EXPECT().
// 					LikeExists(gomock.Any(), 1, 1).
// 					Return(true, nil)

// 				es.EXPECT().
// 					GetUser(gomock.Any(), 1).
// 					Return(&dto.EmployerProfileResponse{
// 						ID:          1,
// 						CompanyName: "Tech Corp",
// 						Slogan:      "Иван",
// 						Website:     "Иванов",
// 						Email:       "ivan@tech.com",
// 					}, nil)
// 			},
// 			expectedResult: []dto.VacancyShortResponse{
// 				{
// 					ID:             1,
// 					Title:          "Senior Backend Developer",
// 					Employer:       &dto.EmployerProfileResponse{ID: 1, CompanyName: "Tech Corp", Slogan: "Иван", Website: "Иванов", Email: "ivan@tech.com"},
// 					Specialization: "Backend разработка",
// 					WorkFormat:     "remote",
// 					Employment:     "full",
// 					WorkingHours:   18,
// 					SalaryFrom:     150000,
// 					SalaryTo:       250000,
// 					TaxesIncluded:  true,
// 					CreatedAt:      now.Format(time.RFC3339),
// 					UpdatedAt:      now.Format(time.RFC3339),
// 					City:           "Москва",
// 					Responded:      false,
// 					Liked:          true,
// 				},
// 			},
// 			expectedErr: nil,
// 		},
// 		{
// 			name:            "Поиск для неавторизованного пользователя",
// 			userID:          0,
// 			userRole:        "",
// 			searchQuery:     "frontend",
// 			specializations: []string{"Frontend разработка"},
// 			limit:           10,
// 			offset:          0,
// 			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
// 				vr.EXPECT().
// 					FindSpecializationIDsByNames(gomock.Any(), []string{"Frontend разработка"}).
// 					Return([]int{2}, nil)

// 				vr.EXPECT().
// 					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "frontend", []int{2}, 10, 0).
// 					Return([]*entity.Vacancy{
// 						{
// 							ID:               2,
// 							Title:            "Frontend Developer",
// 							EmployerID:       2,
// 							SpecializationID: 2,
// 							WorkFormat:       "office",
// 							Employment:       "full",
// 							WorkingHours:     19,
// 							SalaryFrom:       120000,
// 							SalaryTo:         180000,
// 							TaxesIncluded:    false,
// 							City:             "Санкт-Петербург",
// 							CreatedAt:        now,
// 							UpdatedAt:        now,
// 						},
// 					}, nil)

// 				sr.EXPECT().
// 					GetByID(gomock.Any(), 2).
// 					Return(&entity.Specialization{
// 						ID:   2,
// 						Name: "Frontend разработка",
// 					}, nil)

// 				es.EXPECT().
// 					GetUser(gomock.Any(), 2).
// 					Return(&dto.EmployerProfileResponse{
// 						ID:          2,
// 						CompanyName: "Web Inc",
// 						Slogan:      "Петр",
// 						Website:     "Петров",
// 						Email:       "petr@web.com",
// 					}, nil)
// 			},
// 			expectedResult: []dto.VacancyShortResponse{
// 				{
// 					ID:             2,
// 					Title:          "Frontend Developer",
// 					Employer:       &dto.EmployerProfileResponse{ID: 2, CompanyName: "Web Inc", Slogan: "Петр", Website: "Петров", Email: "petr@web.com"},
// 					Specialization: "Frontend разработка",
// 					WorkFormat:     "office",
// 					Employment:     "full",
// 					WorkingHours:   19,
// 					SalaryFrom:     120000,
// 					SalaryTo:       180000,
// 					TaxesIncluded:  false,
// 					CreatedAt:      now.Format(time.RFC3339),
// 					UpdatedAt:      now.Format(time.RFC3339),
// 					City:           "Санкт-Петербург",
// 					Responded:      false,
// 					Liked:          false,
// 				},
// 			},
// 			expectedErr: nil,
// 		},
// 		{
// 			name:            "Ошибка при поиске ID специализаций",
// 			userID:          1,
// 			userRole:        "applicant",
// 			searchQuery:     "devops",
// 			specializations: []string{"DevOps"},
// 			limit:           10,
// 			offset:          0,
// 			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
// 				vr.EXPECT().
// 					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps"}).
// 					Return(nil, entity.NewError(
// 						entity.ErrInternal,
// 						fmt.Errorf("ошибка при поиске специализаций"),
// 					))
// 			},
// 			expectedResult: nil,
// 			expectedErr: entity.NewError(
// 				entity.ErrInternal,
// 				fmt.Errorf("ошибка при поиске специализаций"),
// 			),
// 		},
// 		{
// 			name:            "Не найдено специализаций",
// 			userID:          1,
// 			userRole:        "applicant",
// 			searchQuery:     "design",
// 			specializations: []string{"UI/UX Design"},
// 			limit:           10,
// 			offset:          0,
// 			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
// 				vr.EXPECT().
// 					FindSpecializationIDsByNames(gomock.Any(), []string{"UI/UX Design"}).
// 					Return([]int{}, nil)
// 			},
// 			expectedResult: []dto.VacancyShortResponse{},
// 			expectedErr:    nil,
// 		},
// 		{
// 			name:            "Ошибка при поиске вакансий",
// 			userID:          1,
// 			userRole:        "applicant",
// 			searchQuery:     "backend",
// 			specializations: []string{"Backend разработка"},
// 			limit:           10,
// 			offset:          0,
// 			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
// 				vr.EXPECT().
// 					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
// 					Return([]int{1}, nil)

// 				vr.EXPECT().
// 					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "backend", []int{1}, 10, 0).
// 					Return(nil, entity.NewError(
// 						entity.ErrInternal,
// 						fmt.Errorf("ошибка при поиске вакансий"),
// 					))
// 			},
// 			expectedResult: nil,
// 			expectedErr: entity.NewError(
// 				entity.ErrInternal,
// 				fmt.Errorf("ошибка при поиске вакансий"),
// 			),
// 		},
// 		{
// 			name:            "Ошибка при проверке отклика",
// 			userID:          1,
// 			userRole:        "applicant",
// 			searchQuery:     "backend",
// 			specializations: []string{"Backend разработка"},
// 			limit:           10,
// 			offset:          0,
// 			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
// 				vr.EXPECT().
// 					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
// 					Return([]int{1}, nil)

// 				vr.EXPECT().
// 					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "backend", []int{1}, 10, 0).
// 					Return([]*entity.Vacancy{
// 						{
// 							ID: 1,
// 						},
// 					}, nil)

// 				vr.EXPECT().
// 					ResponseExists(gomock.Any(), 1, 1).
// 					Return(false, entity.NewError(
// 						entity.ErrInternal,
// 						fmt.Errorf("ошибка при проверке отклика"),
// 					))
// 			},
// 			expectedResult: nil,
// 			expectedErr: entity.NewError(
// 				entity.ErrInternal,
// 				fmt.Errorf("ошибка при проверке отклика"),
// 			),
// 		},
// 		{
// 			name:            "Ошибка при проверке лайка",
// 			userID:          1,
// 			userRole:        "applicant",
// 			searchQuery:     "backend",
// 			specializations: []string{"Backend разработка"},
// 			limit:           10,
// 			offset:          0,
// 			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
// 				vr.EXPECT().
// 					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
// 					Return([]int{1}, nil)

// 				vr.EXPECT().
// 					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "backend", []int{1}, 10, 0).
// 					Return([]*entity.Vacancy{
// 						{
// 							ID: 1,
// 						},
// 					}, nil)

// 				vr.EXPECT().
// 					ResponseExists(gomock.Any(), 1, 1).
// 					Return(false, nil)

// 				vr.EXPECT().
// 					LikeExists(gomock.Any(), 1, 1).
// 					Return(false, entity.NewError(
// 						entity.ErrInternal,
// 						fmt.Errorf("ошибка при проверке лайка"),
// 					))
// 			},
// 			expectedResult: nil,
// 			expectedErr: entity.NewError(
// 				entity.ErrInternal,
// 				fmt.Errorf("ошибка при проверке лайка"),
// 			),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			t.Parallel()

// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()

// 			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
// 			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
// 			mockEmployerService := m.NewMockEmployer(ctrl)
// 			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
// 			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
// 			mockApplicantService := m.NewMockApplicant(ctrl)

// 			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

// 			service := NewVacanciesService(
// 				mockVacancyRepo,
// 				mockApplicantRepo,
// 				mockSpecRepo,
// 				mockEmployerService,
// 				mockResumeRepo,
// 				mockApplicantService,
// 			)
// 			ctx := context.Background()

// 			result, err := service.SearchVacanciesByQueryAndSpecializations(
// 				ctx,
// 				tc.userID,
// 				tc.userRole,
// 				tc.searchQuery,
// 				tc.specializations,
// 				tc.limit,
// 				tc.offset,
// 			)

// 			if tc.expectedErr != nil {
// 				require.Error(t, err)
// 				var serviceErr entity.Error
// 				require.ErrorAs(t, err, &serviceErr)
// 				require.Equal(t, tc.expectedErr.Error(), err.Error())
// 			} else {
// 				require.NoError(t, err)
// 				require.Equal(t, tc.expectedResult, result)
// 			}
// 		})
// 	}
// }

func TestVacanciesService_DeleteVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		id             int
		employerID     int
		mockSetup      func(*mock.MockVacancyRepository)
		expectedResult *dto.DeleteVacancy
		expectedErr    error
	}{
		{
			name:       "Успешное удаление вакансии",
			id:         1,
			employerID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:         1,
						EmployerID: 1,
					}, nil)

				vr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				vr.EXPECT().
					DeleteCity(gomock.Any(), 1).
					Return(nil)

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
			name:       "Вакансия не найдена",
			id:         999,
			employerID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("вакансия с id=999 не найдена"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("вакансия с id=999 не найдена"),
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
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
				mockResumeRepo,
				mockApplicantService,
			)
			ctx := context.Background()

			result, err := service.DeleteVacancy(ctx, tc.id, tc.employerID)

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
		resumeID      int
		mockSetup     func(*mock.MockVacancyRepository)
		expectedNotif entity.Notification
		expectedErr   error
		expectedErrAs interface{}
	}{
		{
			name:        "Успешный отклик на вакансию",
			vacancyID:   1,
			applicantID: 1,
			resumeID:    1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1, EmployerID: 2}, nil)

				vr.EXPECT().
					ResponseExistsForApplicant(gomock.Any(), 1, 1, 1).
					Return(false, nil)

				vr.EXPECT().
					CreateResponse(gomock.Any(), 1, 1, 1).
					Return(nil)
			},
			expectedNotif: entity.Notification{
				Type:         entity.ApplyNotificationType,
				SenderID:     1,
				SenderRole:   entity.ApplicantRole,
				ReceiverID:   2,
				ReceiverRole: entity.EmployerRole,
				ObjectID:     1,
				ResumeID:     1,
			},
			expectedErr: nil,
		},
		{
			name:        "Повторный отклик удаляет предыдущий",
			vacancyID:   1,
			applicantID: 1,
			resumeID:    1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1, EmployerID: 2}, nil)

				vr.EXPECT().
					ResponseExistsForApplicant(gomock.Any(), 1, 1, 1).
					Return(true, nil)

				vr.EXPECT().
					DeleteResponse(gomock.Any(), 1, 1, 1).
					Return(nil)
			},
			expectedNotif: entity.Notification{
				Type:         "",
				SenderID:     0,
				SenderRole:   "",
				ReceiverID:   0,
				ReceiverRole: "",
				ObjectID:     0,
				ResumeID:     0,
				IsViewed:     false,
			},
			expectedErr: nil,
		},
		{
			name:        "Вакансия не найдена",
			vacancyID:   999,
			applicantID: 1,
			resumeID:    1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("vacancy not found"),
					))
			},
			expectedNotif: entity.Notification{},
			expectedErr: fmt.Errorf("vacancy not found: %w", entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("vacancy not found"),
			)),
		},
		{
			name:        "Ошибка проверки существующего отклика",
			vacancyID:   1,
			applicantID: 1,
			resumeID:    1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					ResponseExistsForApplicant(gomock.Any(), 1, 1, 1).
					Return(false, fmt.Errorf("database error"))
			},
			expectedNotif: entity.Notification{},
			expectedErr:   fmt.Errorf("failed to check existing responses: %w", fmt.Errorf("database error")),
		},
		{
			name:        "Ошибка при удалении отклика",
			vacancyID:   1,
			applicantID: 1,
			resumeID:    1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					ResponseExistsForApplicant(gomock.Any(), 1, 1, 1).
					Return(true, nil)

				vr.EXPECT().
					DeleteResponse(gomock.Any(), 1, 1, 1).
					Return(fmt.Errorf("delete error"))
			},
			expectedNotif: entity.Notification{},
			expectedErr:   fmt.Errorf("delete error"),
		},
		{
			name:        "Ошибка при создании отклика",
			vacancyID:   1,
			applicantID: 1,
			resumeID:    1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1, EmployerID: 2}, nil)

				vr.EXPECT().
					ResponseExistsForApplicant(gomock.Any(), 1, 1, 1).
					Return(false, nil)

				vr.EXPECT().
					CreateResponse(gomock.Any(), 1, 1, 1).
					Return(fmt.Errorf("create error"))
			},
			expectedNotif: entity.Notification{},
			expectedErr:   fmt.Errorf("create error"),
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
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
				mockResumeRepo,
				mockApplicantService,
			)
			ctx := context.Background()

			notification, err := service.ApplyToVacancy(ctx, tc.vacancyID, tc.applicantID, tc.resumeID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				if tc.expectedErrAs != nil {
					require.ErrorAs(t, err, tc.expectedErrAs)
				} else {
					require.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedNotif, notification)
			}
		})
	}
}

func TestVacanciesService_LikeVacancy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		vacancyID   int
		applicantID int
		mockSetup   func(*mock.MockVacancyRepository)
		expectedErr error
	}{
		{
			name:        "Успешное добавление лайка",
			vacancyID:   1,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(false, nil)

				vr.EXPECT().
					CreateLike(gomock.Any(), 1, 1).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Неуспешное добавление лайка",
			vacancyID:   1,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{}, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(false, nil)

				vr.EXPECT().
					CreateLike(gomock.Any(), 1, 1).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Удаление лайка, если уже лайкнуто",
			vacancyID:   1,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(true, nil)

				vr.EXPECT().
					DeleteLike(gomock.Any(), 1, 1).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Лайк не получилось проверить",
			vacancyID:   1,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(false, fmt.Errorf("failed to check existing like:"))
			},
			expectedErr: fmt.Errorf("failed to check existing like:"),
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
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
				mockResumeRepo,
				mockApplicantService,
			)
			ctx := context.Background()

			err := service.LikeVacancy(ctx, tc.vacancyID, tc.applicantID)

			if tc.expectedErr != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestVacanciesService_GetLikedVacancies(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		applicantID    int
		limit          int
		offset         int
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult []dto.VacancyShortResponse
		expectedErr    error
	}{
		{
			name:        "Успешное получение понравившихся вакансий",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetlikedVacancies(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       1,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full",
							WorkingHours:     18,
							SalaryFrom:       100000,
							SalaryTo:         200000,
							TaxesIncluded:    true,
							City:             "Москва",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.EmployerProfileResponse{
						ID:          1,
						CompanyName: "Tech Corp",
						Slogan:      "Иван",
						Website:     "Иванов",
						Email:       "ivan@tech.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 1, CompanyName: "Tech Corp", Slogan: "Иван", Website: "Иванов", Email: "ivan@tech.com"},
					Specialization: "Backend разработка",
					WorkFormat:     "remote",
					Employment:     "full",
					WorkingHours:   18,
					SalaryFrom:     100000,
					SalaryTo:       200000,
					TaxesIncluded:  true,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Москва",
					Responded:      false,
					Liked:          true,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Успешное получение понравившихся вакансий",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetlikedVacancies(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       1,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full",
							WorkingHours:     18,
							SalaryFrom:       100000,
							SalaryTo:         200000,
							TaxesIncluded:    true,
							City:             "Москва",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(true, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.EmployerProfileResponse{
						ID:          1,
						CompanyName: "Tech Corp",
						Slogan:      "Иван",
						Website:     "Иванов",
						Email:       "ivan@tech.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 1, CompanyName: "Tech Corp", Slogan: "Иван", Website: "Иванов", Email: "ivan@tech.com"},
					Specialization: "Backend разработка",
					WorkFormat:     "remote",
					Employment:     "full",
					WorkingHours:   18,
					SalaryFrom:     100000,
					SalaryTo:       200000,
					TaxesIncluded:  true,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Москва",
					Responded:      true,
					Liked:          true,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка при получении списка вакансий",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetlikedVacancies(gomock.Any(), 1, 10, 0).
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
			name:        "Ошибка при получении специализации (пропускаем вакансию)",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetlikedVacancies(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
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
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:        "Ошибка при получении информации о работодателе (пропускаем вакансию)",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetlikedVacancies(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:         1,
							EmployerID: 1,
						},
					}, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении информации о работодателе"),
					))
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:        "Ошибка при проверке отклика",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				// Мок для получения списка вакансий
				vr.EXPECT().
					GetlikedVacancies(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							EmployerID:       1,
							SpecializationID: 1,
						},
					}, nil)

				// Мок для получения специализации
				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				// Мок для получения информации о работодателе
				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.EmployerProfileResponse{
						ID:          1,
						CompanyName: "Test Company",
					}, nil)

				// Мок для проверки отклика (должен вернуть ошибку)
				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
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
			name:        "Пустой список понравившихся вакансий",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetlikedVacancies(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
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
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
				mockResumeRepo,
				mockApplicantService,
			)
			ctx := context.Background()

			result, err := service.GetLikedVacancies(ctx, tc.applicantID, tc.limit, tc.offset)

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

func TestVacanciesService_GetRespondedResumeOnVacancy(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name      string
		vacancyID int
		limit     int
		offset    int
		mockSetup func(
			vr *mock.MockVacancyRepository,
			rr *mock.MockResumeRepository,
			sr *mock.MockSpecializationRepository,
			as *m.MockApplicant,
		)
		expectedResult []dto.ResumeShortResponse
		expectedErr    error
	}{
		{
			name:      "Успешное получение списка резюме",
			vacancyID: 1,
			limit:     10,
			offset:    0,
			mockSetup: func(
				vr *mock.MockVacancyRepository,
				rr *mock.MockResumeRepository,
				sr *mock.MockSpecializationRepository,
				as *m.MockApplicant,
			) {
				vr.EXPECT().
					GetVacancyResponses(gomock.Any(), 1, 10, 0).
					Return([]*entity.VacancyResponses{
						{
							ID:          1,
							VacancyID:   1,
							ApplicantID: 1,
							ResumeID:    1,
							AppliedAt:   now,
						},
					}, nil)

				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:               1,
						ApplicantID:      1,
						SpecializationID: 2,
						Profession:       "Developer",
						CreatedAt:        now,
						UpdatedAt:        now,
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{
						ID:   2,
						Name: "Backend разработка",
					}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return([]entity.WorkExperience{
						{
							ID:           1,
							EmployerName: "Company",
							Position:     "Senior Dev",
							Duties:       "Coding",
							Achievements: "Awards",
							StartDate:    now.AddDate(-1, 0, 0),
							UntilNow:     false,
							EndDate:      now,
						},
					}, nil)

				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{
						ID:        1,
						FirstName: "Иван",
						LastName:  "Иванов",
					}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             1,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "Backend разработка",
					Profession:     "Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					WorkExperience: dto.WorkExperienceShort{
						ID:           1,
						EmployerName: "Company",
						Position:     "Senior Dev",
						Duties:       "Coding",
						Achievements: "Awards",
						StartDate:    now.AddDate(-1, 0, 0).Format("2006-01-02"),
						UntilNow:     false,
						EndDate:      now.Format("2006-01-02"),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:      "Ошибка при получении списка откликов",
			vacancyID: 1,
			limit:     10,
			offset:    0,
			mockSetup: func(
				vr *mock.MockVacancyRepository,
				rr *mock.MockResumeRepository,
				sr *mock.MockSpecializationRepository,
				as *m.MockApplicant,
			) {
				vr.EXPECT().
					GetVacancyResponses(gomock.Any(), 1, 10, 0).
					Return(nil, fmt.Errorf("db error"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("failed to get vacancy responses: db error"),
		},
		{
			name:      "Ошибка при получении резюме (пропуск резюме)",
			vacancyID: 1,
			limit:     10,
			offset:    0,
			mockSetup: func(
				vr *mock.MockVacancyRepository,
				rr *mock.MockResumeRepository,
				sr *mock.MockSpecializationRepository,
				as *m.MockApplicant,
			) {
				vr.EXPECT().
					GetVacancyResponses(gomock.Any(), 1, 10, 0).
					Return([]*entity.VacancyResponses{
						{ResumeID: 1},
					}, nil)

				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, fmt.Errorf("not found"))
			},
			expectedResult: []dto.ResumeShortResponse{},
			expectedErr:    fmt.Errorf("not found"),
		},
		// Можно добавить больше кейсов по аналогии (например, ошибки при получении специализации, опыта, пользователя)
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo, mockResumeRepo, mockSpecRepo, mockApplicantService)

			service := NewVacanciesService(
				mockVacancyRepo,
				nil, // applicantRepo not used here
				mockSpecRepo,
				nil, // employerService not used here
				mockResumeRepo,
				mockApplicantService,
			)

			ctx := context.Background()

			resumes, err := service.GetRespondedResumeOnVacancy(ctx, tc.vacancyID, tc.limit, tc.offset)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, resumes)
			}
		})
	}
}

func TestVacanciesService_SearchVacanciesBySpecializations(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name            string
		userID          int
		userRole        string
		specializations []string
		limit           int
		offset          int
		mockSetup       func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult  []dto.VacancyShortResponse
		expectedErr     error
	}{
		{
			name:            "Успешный поиск по специализациям для соискателя",
			userID:          1,
			userRole:        "applicant",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesBySpecializations(gomock.Any(), []int{1}, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							Title:            "Senior Backend Developer",
							EmployerID:       1,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full_time",
							WorkingHours:     19,
							SalaryFrom:       150000,
							SalaryTo:         250000,
							TaxesIncluded:    true,
							City:             "Москва",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(true, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.EmployerProfileResponse{
						ID:          1,
						CompanyName: "Tech Corp",
						Slogan:      "Иван",
						Website:     "Иванов",
						Email:       "ivan@tech.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Senior Backend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 1, CompanyName: "Tech Corp", Slogan: "Иван", Website: "Иванов", Email: "ivan@tech.com"},
					Specialization: "Backend разработка",
					WorkFormat:     "remote",
					Employment:     "full_time",
					WorkingHours:   19,
					SalaryFrom:     150000,
					SalaryTo:       250000,
					TaxesIncluded:  true,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Москва",
					Responded:      false,
					Liked:          true,
				},
			},
			expectedErr: nil,
		},
		{
			name:            "Поиск для неавторизованного пользователя",
			userID:          0,
			userRole:        "",
			specializations: []string{"Frontend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Frontend разработка"}).
					Return([]int{2}, nil)

				vr.EXPECT().
					SearchVacanciesBySpecializations(gomock.Any(), []int{2}, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               2,
							Title:            "Frontend Developer",
							EmployerID:       2,
							SpecializationID: 2,
							WorkFormat:       "office",
							Employment:       "full_time",
							WorkingHours:     19,
							SalaryFrom:       120000,
							SalaryTo:         180000,
							TaxesIncluded:    false,
							City:             "Санкт-Петербург",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{
						ID:   2,
						Name: "Frontend разработка",
					}, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.EmployerProfileResponse{
						ID:          2,
						CompanyName: "Web Inc",
						Slogan:      "Петр",
						Website:     "Петров",
						Email:       "petr@web.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             2,
					Title:          "Frontend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 2, CompanyName: "Web Inc", Slogan: "Петр", Website: "Петров", Email: "petr@web.com"},
					Specialization: "Frontend разработка",
					WorkFormat:     "office",
					Employment:     "full_time",
					WorkingHours:   19,
					SalaryFrom:     120000,
					SalaryTo:       180000,
					TaxesIncluded:  false,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Санкт-Петербург",
					Responded:      false,
					Liked:          false,
				},
			},
			expectedErr: nil,
		},
		{
			name:            "Ошибка при поиске ID специализаций",
			userID:          1,
			userRole:        "applicant",
			specializations: []string{"DevOps"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps"}).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при поиске специализаций"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске специализаций"),
			),
		},
		{
			name:            "Не найдено специализаций",
			userID:          1,
			userRole:        "applicant",
			specializations: []string{"UI/UX Design"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"UI/UX Design"}).
					Return([]int{}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:            "Ошибка при поиске вакансий",
			userID:          1,
			userRole:        "applicant",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesBySpecializations(gomock.Any(), []int{1}, 10, 0).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при поиске вакансий"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске вакансий"),
			),
		},
		{
			name:            "Ошибка при получении специализации (пропускаем вакансию)",
			userID:          1,
			userRole:        "applicant",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesBySpecializations(gomock.Any(), []int{1}, 10, 0).
					Return([]*entity.Vacancy{
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

				es.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)

				vr.EXPECT().
					ResponseExists(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)

				vr.EXPECT().
					LikeExists(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		// {
		// 	name:            "Ошибка при получении специализации (пропускаем вакансию)",
		// 	userID:          1,
		// 	userRole:        "applicant",
		// 	specializations: []string{"Frontend разработка"},
		// 	limit:           10,
		// 	offset:          0,
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
		// 		vr.EXPECT().
		// 			FindSpecializationIDsByNames(gomock.Any(), []string{"Frontend разработка"}).
		// 			Return([]int{2}, nil)

		// 		vr.EXPECT().
		// 			SearchVacanciesBySpecializations(gomock.Any(), []int{2}, 10, 0).
		// 			Return([]*entity.Vacancy{
		// 				{
		// 					ID:               2,
		// 					Title:            "Frontend Developer",
		// 					EmployerID:       2,
		// 					SpecializationID: 2,
		// 					WorkFormat:       "office",
		// 					Employment:       "full_time",
		// 					WorkingHours:     19,
		// 					SalaryFrom:       120000,
		// 					SalaryTo:         180000,
		// 					TaxesIncluded:    false,
		// 					Experience:       "1_3_years",
		// 					City:             "Санкт-Петербург",
		// 					Schedule:         "5/2",
		// 					CreatedAt:        now,
		// 					UpdatedAt:        now,
		// 				},
		// 			}, nil)

		// 		es.EXPECT().
		// 			GetUser(gomock.Any(), gomock.Any()).
		// 			Return(nil, entity.NewError(
		// 				entity.ErrInternal,
		// 				fmt.Errorf("ошибка при конвертации работодателя в DTO"),
		// 			))

		// 		vr.EXPECT().
		// 			ResponseExists(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)

		// 		vr.EXPECT().
		// 			LikeExists(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	expectedResult: []dto.VacancyShortResponse{},
		// 	expectedErr: entity.NewError(
		// 		entity.ErrInternal,
		// 		fmt.Errorf("ошибка при конвертации работодателя в DTO"),
		// 	),
		// },
		{
			name:            "Ошибка при проверке отклика",
			userID:          1,
			userRole:        "applicant",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesBySpecializations(gomock.Any(), []int{1}, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID: 1,
						},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
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
			name:            "Ошибка при проверке лайка",
			userID:          1,
			userRole:        "applicant",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesBySpecializations(gomock.Any(), []int{1}, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID: 1,
						},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(false, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при проверке лайка"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке лайка"),
			),
		},
		// {
		// 	name:            "Ошибка при получении информации о работодателе (пропускаем вакансию)",
		// 	userID:          1,
		// 	userRole:        "applicant",
		// 	specializations: []string{"Backend разработка"},
		// 	limit:           10,
		// 	offset:          0,
		// 	mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
		// 		vr.EXPECT().
		// 			FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
		// 			Return([]int{1}, nil)

		// 		vr.EXPECT().
		// 			SearchVacanciesBySpecializations(gomock.Any(), []int{1}, 10, 0).
		// 			Return([]*entity.Vacancy{
		// 				{
		// 					ID:               1,
		// 					Title:            "Frontend Developer",
		// 					EmployerID:       1,
		// 					SpecializationID: 1,
		// 					WorkFormat:       "office",
		// 					Employment:       "full_time",
		// 					WorkingHours:     19,
		// 					SalaryFrom:       120000,
		// 					SalaryTo:         180000,
		// 					TaxesIncluded:    false,
		// 					City:             "Санкт-Петербург",
		// 					CreatedAt:        now,
		// 					UpdatedAt:        now,
		// 				},
		// 			}, nil)

		// 		es.EXPECT().
		// 			GetUser(gomock.Any(), 1).
		// 			Times(0)

		// 		vr.EXPECT().
		// 			ResponseExists(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)

		// 		vr.EXPECT().
		// 			LikeExists(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		// 	expectedResult: []dto.VacancyShortResponse{},
		// 	expectedErr:    nil,
		// },
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockVacancyRepo := mock.NewMockVacancyRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
				mockResumeRepo,
				mockApplicantService,
			)
			ctx := context.Background()

			result, err := service.SearchVacanciesBySpecializations(
				ctx,
				tc.userID,
				tc.userRole,
				tc.specializations,
				tc.limit,
				tc.offset,
			)

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
		userRole       string
		limit          int
		offset         int
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult []dto.VacancyShortResponse
		expectedErr    error
	}{
		{
			name:          "Успешное получение всех вакансий для соискателя",
			currentUserID: 1,
			userRole:      "applicant",
			limit:         10,
			offset:        0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       1,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full",
							WorkingHours:     18,
							SalaryFrom:       100000,
							SalaryTo:         200000,
							TaxesIncluded:    true,
							City:             "Москва",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(true, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.EmployerProfileResponse{
						ID:          1,
						CompanyName: "Tech Corp",
						Slogan:      "Иван",
						Website:     "Иванов",
						Email:       "ivan@tech.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 1, CompanyName: "Tech Corp", Slogan: "Иван", Website: "Иванов", Email: "ivan@tech.com"},
					Specialization: "Backend разработка",
					WorkFormat:     "remote",
					Employment:     "full",
					WorkingHours:   18,
					SalaryFrom:     100000,
					SalaryTo:       200000,
					TaxesIncluded:  true,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Москва",
					Responded:      false,
					Liked:          true,
				},
			},
			expectedErr: nil,
		},
		{
			name:          "Успешное получение всех вакансий для неавторизованного пользователя",
			currentUserID: 0,
			userRole:      "",
			limit:         10,
			offset:        0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               2,
							Title:            "Frontend Developer",
							EmployerID:       2,
							SpecializationID: 2,
							WorkFormat:       "office",
							Employment:       "full",
							WorkingHours:     19,
							SalaryFrom:       120000,
							SalaryTo:         180000,
							TaxesIncluded:    false,
							City:             "Санкт-Петербург",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{
						ID:   2,
						Name: "Frontend разработка",
					}, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.EmployerProfileResponse{
						ID:          2,
						CompanyName: "Web Inc",
						Slogan:      "Петр",
						Website:     "Петров",
						Email:       "petr@web.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             2,
					Title:          "Frontend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 2, CompanyName: "Web Inc", Slogan: "Петр", Website: "Петров", Email: "petr@web.com"},
					Specialization: "Frontend разработка",
					WorkFormat:     "office",
					Employment:     "full",
					WorkingHours:   19,
					SalaryFrom:     120000,
					SalaryTo:       180000,
					TaxesIncluded:  false,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Санкт-Петербург",
					Responded:      false,
					Liked:          false,
				},
			},
			expectedErr: nil,
		},
		{
			name:          "Ошибка при получении списка вакансий",
			currentUserID: 0,
			userRole:      "",
			limit:         10,
			offset:        0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
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
			userRole:      "",
			limit:         10,
			offset:        0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]*entity.Vacancy{
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
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:          "Ошибка при проверке отклика",
			currentUserID: 1,
			userRole:      "applicant",
			limit:         10,
			offset:        0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]*entity.Vacancy{
						{
							ID: 1,
						},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
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
			name:          "Ошибка при проверке лайка",
			currentUserID: 1,
			userRole:      "applicant",
			limit:         10,
			offset:        0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]*entity.Vacancy{
						{
							ID: 1,
						},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(false, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при проверке лайка"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке лайка"),
			),
		},
		{
			name:          "Ошибка при получении информации о работодателе (пропускаем вакансию)",
			currentUserID: 0,
			userRole:      "",
			limit:         10,
			offset:        0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:         1,
							EmployerID: 1,
						},
					}, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении информации о работодателе"),
					))
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:          "Пустой список вакансий",
			currentUserID: 0,
			userRole:      "",
			limit:         10,
			offset:        0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]*entity.Vacancy{}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
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
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
				mockResumeRepo,
				mockApplicantService,
			)
			ctx := context.Background()

			result, err := service.GetAll(ctx, tc.currentUserID, tc.userRole, tc.limit, tc.offset)

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

func TestVacanciesService_GetActiveVacanciesByEmployerID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		employerID     int
		userID         int
		userRole       string
		limit          int
		offset         int
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult []dto.VacancyShortResponse
		expectedErr    error
	}{
		{
			name:       "Успешное получение активных вакансий для работодателя",
			employerID: 1,
			userID:     2,
			userRole:   "applicant",
			limit:      10,
			offset:     0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							EmployerID:       1,
							Title:            "Backend Developer",
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full",
							WorkingHours:     18,
							SalaryFrom:       100000,
							SalaryTo:         150000,
							TaxesIncluded:    true,
							City:             "Москва",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 2).
					Return(false, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 2).
					Return(true, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.EmployerProfileResponse{
						ID:          1,
						CompanyName: "Tech Corp",
						Email:       "employer@example.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 1, CompanyName: "Tech Corp", Email: "employer@example.com"},
					Specialization: "Backend разработка",
					WorkFormat:     "remote",
					Employment:     "full",
					WorkingHours:   18,
					SalaryFrom:     100000,
					SalaryTo:       150000,
					TaxesIncluded:  true,
					City:           "Москва",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					Responded:      false,
					Liked:          true,
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Вакансия без специализации",
			employerID: 1,
			userID:     0,
			userRole:   "",
			limit:      10,
			offset:     0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:            2,
							EmployerID:    1,
							Title:         "Frontend Developer",
							WorkFormat:    "hybrid",
							Employment:    "part",
							WorkingHours:  19,
							SalaryFrom:    80000,
							SalaryTo:      120000,
							TaxesIncluded: false,
							City:          "Санкт-Петербург",
							CreatedAt:     now,
							UpdatedAt:     now,
						},
					}, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.EmployerProfileResponse{
						ID:          1,
						CompanyName: "Tech Corp",
						Email:       "employer@example.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:            2,
					Title:         "Frontend Developer",
					Employer:      &dto.EmployerProfileResponse{ID: 1, CompanyName: "Tech Corp", Email: "employer@example.com"},
					WorkFormat:    "hybrid",
					Employment:    "part",
					WorkingHours:  19,
					SalaryFrom:    80000,
					SalaryTo:      120000,
					TaxesIncluded: false,
					City:          "Санкт-Петербург",
					CreatedAt:     now.Format(time.RFC3339),
					UpdatedAt:     now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Ошибка при получении вакансий",
			employerID: 1,
			userID:     0,
			userRole:   "",
			limit:      10,
			offset:     0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 1, 10, 0).
					Return(nil, fmt.Errorf("ошибка базы данных"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка базы данных"),
		},
		{
			name:       "Ошибка при получении специализации (пропускаем вакансию)",
			employerID: 1,
			userID:     0,
			userRole:   "",
			limit:      10,
			offset:     0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetActiveVacanciesByEmployerID(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               3,
							EmployerID:       1,
							Title:            "DevOps Engineer",
							SpecializationID: 2,
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(nil, fmt.Errorf("ошибка при получении специализации"))
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
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
			mockEmployerService := m.NewMockEmployer(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := &VacanciesService{
				vacanciesRepository:      mockVacancyRepo,
				specializationRepository: mockSpecRepo,
				employerService:          mockEmployerService,
			}

			ctx := context.Background()
			result, err := service.GetActiveVacanciesByEmployerID(ctx, tc.employerID, tc.userID, tc.userRole, tc.limit, tc.offset)

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
		limit          int
		offset         int
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult []dto.VacancyShortResponse
		expectedErr    error
	}{
		{
			name:        "Успешное получение вакансий с заполненными полями",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							EmployerID:       2,
							Title:            "Backend Developer",
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full",
							WorkingHours:     40,
							SalaryFrom:       120000,
							SalaryTo:         150000,
							TaxesIncluded:    true,
							City:             "Москва",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(true, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(false, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.EmployerProfileResponse{
						ID:          2,
						CompanyName: "Tech Corp",
						Email:       "employer@example.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 2, CompanyName: "Tech Corp", Email: "employer@example.com"},
					Specialization: "Backend разработка",
					WorkFormat:     "remote",
					Employment:     "full",
					WorkingHours:   40,
					SalaryFrom:     120000,
					SalaryTo:       150000,
					TaxesIncluded:  true,
					City:           "Москва",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					Responded:      true,
					Liked:          false,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Вакансия без специализации (пропускается)",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{ID: 3, EmployerID: 1, Title: "DevOps Engineer", SpecializationID: 99, CreatedAt: now, UpdatedAt: now},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 99).
					Return(nil, errors.New("not found"))
				// остальные вызовы не должны быть сделаны
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
		{
			name:        "Ошибка при проверке отклика",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{ID: 1, EmployerID: 2, Title: "Backend Developer"},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, fmt.Errorf("ошибка базы данных"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка базы данных"),
		},
		{
			name:        "Ошибка при получении вакансий",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 1, 10, 0).
					Return(nil, fmt.Errorf("ошибка базы данных"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка базы данных"),
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
			mockEmployerService := m.NewMockEmployer(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := &VacanciesService{
				vacanciesRepository:      mockVacancyRepo,
				specializationRepository: mockSpecRepo,
				employerService:          mockEmployerService,
			}

			ctx := context.Background()
			result, err := service.GetVacanciesByApplicantID(ctx, tc.applicantID, tc.limit, tc.offset)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestVacanciesService_SearchVacancies(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		name           string
		userID         int
		userRole       string
		searchQuery    string
		limit          int
		offset         int
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult []dto.VacancyShortResponse
		expectedErr    error
	}{
		{
			name:        "Успешный поиск вакансий для соискателя",
			userID:      1,
			userRole:    "applicant",
			searchQuery: "developer",
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					SearchVacancies(gomock.Any(), "developer", 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							EmployerID:       2,
							Title:            "Backend Developer",
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full",
							WorkingHours:     18,
							SalaryFrom:       100000,
							SalaryTo:         150000,
							TaxesIncluded:    true,
							City:             "Москва",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)

				vr.EXPECT().
					LikeExists(gomock.Any(), 1, 1).
					Return(true, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.EmployerProfileResponse{
						ID:          2,
						CompanyName: "Tech Corp",
						Email:       "employer@example.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Title:          "Backend Developer",
					Employer:       &dto.EmployerProfileResponse{ID: 2, CompanyName: "Tech Corp", Email: "employer@example.com"},
					Specialization: "Backend разработка",
					WorkFormat:     "remote",
					Employment:     "full",
					WorkingHours:   18,
					SalaryFrom:     100000,
					SalaryTo:       150000,
					TaxesIncluded:  true,
					City:           "Москва",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					Responded:      false,
					Liked:          true,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Успешный поиск вакансий для работодателя",
			userID:      2,
			userRole:    "employer",
			searchQuery: "developer",
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					SearchVacanciesByEmployerID(gomock.Any(), 2, "developer", 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							EmployerID:       2,
							Title:            "Backend Developer",
							SpecializationID: 1,
							WorkFormat:       "office",
							Employment:       "5/2",
							WorkingHours:     15,
							SalaryFrom:       100000,
							SalaryTo:         200000,
							TaxesIncluded:    false,
							City:             "Moskow",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// vr.EXPECT().
				// 	LikeExists(gomock.Any(), 1, 1).
				// 	Return(false, nil)

				// vr.EXPECT().
				// 	ResponseExists(gomock.Any(), 1, 1).
				// 	Return(false, nil)

				es.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.EmployerProfileResponse{
						ID:          2,
						CompanyName: "Tech Corp",
						Email:       "employer@example.com",
						Slogan:      "Aaaaaaaaaa",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:    1,
					Title: "Backend Developer",
					Employer: &dto.EmployerProfileResponse{
						ID:          2,
						CompanyName: "Tech Corp",
						Email:       "employer@example.com",
						Slogan:      "Aaaaaaaaaa",
					},
					Specialization: "Backend разработка",
					WorkFormat:     "office",
					Employment:     "5/2",
					WorkingHours:   15,
					SalaryFrom:     100000,
					SalaryTo:       200000,
					TaxesIncluded:  false,
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Moskow",
					Responded:      false,
					Liked:          false,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка при поиске вакансий",
			userID:      1,
			userRole:    "applicant",
			searchQuery: "developer",
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					SearchVacancies(gomock.Any(), "developer", 10, 0).
					Return(nil, fmt.Errorf("ошибка базы данных"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка базы данных"),
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
			mockEmployerService := m.NewMockEmployer(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := &VacanciesService{
				vacanciesRepository:      mockVacancyRepo,
				specializationRepository: mockSpecRepo,
				employerService:          mockEmployerService,
			}

			ctx := context.Background()
			result, err := service.SearchVacancies(ctx, tc.userID, tc.userRole, tc.searchQuery, tc.limit, tc.offset)

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
