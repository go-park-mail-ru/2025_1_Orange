package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	m "ResuMatch/internal/usecase/mock"
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
			name:       "Успешное создание вакансии со специализацией и навыками",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "Backend Developer",
				Specialization: "Backend разработка",
				WorkFormat:     "remote",
				Skills:         []string{"Go", "SQL"},
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Поиск ID специализации
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				// Создание вакансии
				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{
						ID:               1,
						Title:            "Backend Developer",
						EmployerID:       1,
						SpecializationID: 1,
						WorkFormat:       "remote",
						CreatedAt:        now,
						UpdatedAt:        now,
					}, nil)

				// Поиск ID навыков
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
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

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
				Specialization: "Backend разработка",
				WorkFormat:     "remote",
				Skills:         []string{"Go", "SQL"},
				CreatedAt:      now.Format(time.RFC3339),
				UpdatedAt:      now.Format(time.RFC3339),
			},
			expectedErr: nil,
		},
		{
			name:       "Успешное создание вакансии без специализации и навыков",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:      "Frontend Developer",
				WorkFormat: "hybrid",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				// Создание вакансии
				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{
						ID:         2,
						Title:      "Frontend Developer",
						EmployerID: 1,
						WorkFormat: "hybrid",
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)

				// Получение навыков (пустой список)
				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:         2,
				EmployerID: 1,
				Title:      "Frontend Developer",
				WorkFormat: "hybrid",
				Skills:     []string{},
				CreatedAt:  now.Format(time.RFC3339),
				UpdatedAt:  now.Format(time.RFC3339),
			},
			expectedErr: nil,
		},
		{
			name:       "Ошибка при создании вакансии",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title: "Invalid Vacancy",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("ошибка создания вакансии"))
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка создания вакансии"),
		},
		{
			name:       "Ошибка при добавлении навыков",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:  "Backend Developer",
				Skills: []string{"Go"},
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{
						ID:         3,
						Title:      "Backend Developer",
						EmployerID: 1,
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)

				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					AddSkills(gomock.Any(), 3, []int{1}).
					Return(fmt.Errorf("ошибка добавления навыков"))

				// Получение навыков (должно вернуть пустой список, так как добавление не удалось)
				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 3).
					Return([]entity.Skill{}, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:         3,
				EmployerID: 1,
				Title:      "Backend Developer",
				Skills:     []string{},
				CreatedAt:  now.Format(time.RFC3339),
				UpdatedAt:  now.Format(time.RFC3339),
			},
			expectedErr: nil,
		},
		{
			name:       "Ошибка при получении специализации",
			employerID: 1,
			request: &dto.VacancyCreate{
				Title:          "DevOps Engineer",
				Specialization: "DevOps",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "DevOps").
					Return(2, nil)

				vr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Vacancy{
						ID:               4,
						Title:            "DevOps Engineer",
						EmployerID:       1,
						SpecializationID: 2,
						CreatedAt:        now,
						UpdatedAt:        now,
					}, nil)

				// Получение специализации (ошибка)
				sr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(nil, fmt.Errorf("ошибка получения специализации"))

				// Получение навыков (пустой список)
				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 4).
					Return([]entity.Skill{}, nil)
			},
			expectedResult: nil,
			expectedErr:    fmt.Errorf("ошибка получения специализации"),
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
			name:          "Вакансия не найдена",
			id:            999,
			currentUserID: 0,
			userRole:      "",
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(0, entity.NewError(
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

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
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
		mockSetup      func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult *dto.VacancyResponse
		expectedErr    error
	}{
		{
			name:       "Успешное обновление вакансии",
			id:         1,
			employerID: 1,
			request: &dto.VacancyUpdate{
				Title:                "Updated Backend Developer",
				Specialization:       "Backend разработка",
				WorkFormat:           "hybrid",
				Employment:           "part",
				Schedule:             "fixed",
				WorkingHours:         19,
				SalaryFrom:           120000,
				SalaryTo:             220000,
				TaxesIncluded:        false,
				Experience:           "5+ years",
				Description:          "Обновленное описание",
				Tasks:                "Обновленные задачи",
				Requirements:         "Обновленные требования",
				OptionalRequirements: "Обновленные доп. требования",
				Skills:               []string{"Go", "PostgreSQL"},
				City:                 "Москва",
			},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:         1,
						EmployerID: 1,
					}, nil)

				vr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				vr.EXPECT().
					Update(gomock.Any(), &entity.Vacancy{
						ID:                   1,
						EmployerID:           1,
						Title:                "Updated Backend Developer",
						SpecializationID:     1,
						WorkFormat:           "hybrid",
						Employment:           "part",
						Schedule:             "fixed",
						WorkingHours:         19,
						SalaryFrom:           120000,
						SalaryTo:             220000,
						TaxesIncluded:        false,
						Experience:           "5+ years",
						Description:          "Обновленное описание",
						Tasks:                "Обновленные задачи",
						Requirements:         "Обновленные требования",
						OptionalRequirements: "Обновленные доп. требования",
					}).
					Return(&entity.Vacancy{
						ID:                   1,
						EmployerID:           1,
						Title:                "Updated Backend Developer",
						SpecializationID:     1,
						WorkFormat:           "hybrid",
						Employment:           "part",
						Schedule:             "fixed",
						WorkingHours:         19,
						SalaryFrom:           120000,
						SalaryTo:             220000,
						TaxesIncluded:        false,
						Experience:           "5+ years",
						Description:          "Обновленное описание",
						Tasks:                "Обновленные задачи",
						Requirements:         "Обновленные требования",
						OptionalRequirements: "Обновленные доп. требования",
						CreatedAt:            now,
						UpdatedAt:            now,
					}, nil)

				vr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				vr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "PostgreSQL"}).
					Return([]int{1, 3}, nil)

				vr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 3}).
					Return(nil)

				vr.EXPECT().
					DeleteCity(gomock.Any(), 1).
					Return(nil)

				vr.EXPECT().
					FindCityIDsByNames(gomock.Any(), []string{"Москва", "Санкт-Петербург"}).
					Return([]int{1, 2}, nil)

				vr.EXPECT().
					AddCity(gomock.Any(), 1, []int{1, 2}).
					Return(nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				vr.EXPECT().
					GetSkillsByVacancyID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 3, Name: "PostgreSQL"},
					}, nil)

				vr.EXPECT().
					GetCityByVacancyID(gomock.Any(), 1).
					Return([]entity.City{
						{ID: 1, Name: "Москва"},
						{ID: 2, Name: "Санкт-Петербург"},
					}, nil)
			},
			expectedResult: &dto.VacancyResponse{
				ID:                   1,
				EmployerID:           1,
				Title:                "Updated Backend Developer",
				Specialization:       "Backend разработка",
				WorkFormat:           "hybrid",
				Employment:           "part",
				Schedule:             "fixed",
				WorkingHours:         19,
				SalaryFrom:           120000,
				SalaryTo:             220000,
				TaxesIncluded:        false,
				Experience:           "5+ years",
				Description:          "Обновленное описание",
				Tasks:                "Обновленные задачи",
				Requirements:         "Обновленные требования",
				OptionalRequirements: "Обновленные доп. требования",
				Skills:               []string{"Go", "PostgreSQL"},
				CreatedAt:            now.Format(time.RFC3339),
				UpdatedAt:            now.Format(time.RFC3339),
			},
			expectedErr: nil,
		},
		{
			name:       "Вакансия не принадлежит работодателю",
			id:         1,
			employerID: 2,
			request:    &dto.VacancyUpdate{},
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{
						ID:         1,
						EmployerID: 1,
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
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			mockEmployerService := m.NewMockEmployer(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
			)
			ctx := context.Background()

			result, err := service.UpdateVacancy(ctx, tc.id, tc.employerID, tc.request)

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

func TestVacanciesService_SearchVacanciesByQueryAndSpecializations(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name            string
		userID          int
		userRole        string
		searchQuery     string
		specializations []string
		limit           int
		offset          int
		mockSetup       func(*mock.MockVacancyRepository, *mock.MockSpecializationRepository, *m.MockEmployer)
		expectedResult  []dto.VacancyShortResponse
		expectedErr     error
	}{
		{
			name:            "Успешный поиск по запросу и специализациям для соискателя",
			userID:          1,
			userRole:        "applicant",
			searchQuery:     "backend",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "backend", []int{1}, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:               1,
							Title:            "Senior Backend Developer",
							EmployerID:       1,
							SpecializationID: 1,
							WorkFormat:       "remote",
							Employment:       "full",
							WorkingHours:     18,
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
					Employment:     "full",
					WorkingHours:   18,
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
			searchQuery:     "frontend",
			specializations: []string{"Frontend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Frontend разработка"}).
					Return([]int{2}, nil)

				vr.EXPECT().
					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "frontend", []int{2}, 10, 0).
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
			name:            "Ошибка при поиске ID специализаций",
			userID:          1,
			userRole:        "applicant",
			searchQuery:     "devops",
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
			searchQuery:     "design",
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
			searchQuery:     "backend",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "backend", []int{1}, 10, 0).
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
			name:            "Ошибка при проверке отклика",
			userID:          1,
			userRole:        "applicant",
			searchQuery:     "backend",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "backend", []int{1}, 10, 0).
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
			searchQuery:     "backend",
			specializations: []string{"Backend разработка"},
			limit:           10,
			offset:          0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"Backend разработка"}).
					Return([]int{1}, nil)

				vr.EXPECT().
					SearchVacanciesByQueryAndSpecializations(gomock.Any(), "backend", []int{1}, 10, 0).
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

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
			)
			ctx := context.Background()

			result, err := service.SearchVacanciesByQueryAndSpecializations(
				ctx,
				tc.userID,
				tc.userRole,
				tc.searchQuery,
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

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
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
		name        string
		vacancyID   int
		applicantID int
		mockSetup   func(*mock.MockVacancyRepository)
		expectedErr error
	}{
		{
			name:        "Успешный отклик на вакансию",
			vacancyID:   1,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Vacancy{ID: 1}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, nil)

				vr.EXPECT().
					CreateResponse(gomock.Any(), 1, 1).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:        "Вакансия не найдена",
			vacancyID:   999,
			applicantID: 1,
			mockSetup: func(vr *mock.MockVacancyRepository) {
				vr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("vacancy not found"),
					))
			},
			expectedErr: fmt.Errorf("vacancy not found"),
		},
		{
			name:        "Уже откликался на вакансию",
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
			expectedErr: entity.NewError(entity.ErrAlreadyExists,
				fmt.Errorf("you have already applied to this vacancy")),
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

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
			)
			ctx := context.Background()

			_, err := service.ApplyToVacancy(ctx, tc.vacancyID, tc.applicantID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				if entityErr, ok := tc.expectedErr.(entity.Error); ok {
					var serviceErr entity.Error
					require.ErrorAs(t, err, &serviceErr)
					require.Equal(t, entityErr.Error(), err.Error())
				} else {
					require.Equal(t, tc.expectedErr.Error(), err.Error())
				}
			} else {
				require.NoError(t, err)
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

			tc.mockSetup(mockVacancyRepo)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
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
				vr.EXPECT().
					GetlikedVacancies(gomock.Any(), 1, 10, 0).
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

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
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
							Employment:       "full",
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
					Employment:     "full",
					WorkingHours:   18,
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
					Return(&dto.EmployerProfileResponse{}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{},
			expectedErr:    nil,
		},
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
		{
			name:            "Ошибка при получении информации о работодателе (пропускаем вакансию)",
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

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
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

			tc.mockSetup(mockVacancyRepo, mockSpecRepo, mockEmployerService)

			service := NewVacanciesService(
				mockVacancyRepo,
				mockApplicantRepo,
				mockSpecRepo,
				mockEmployerService,
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
					ID:        3,
					Title:     "DevOps Engineer",
					Employer:  &dto.EmployerProfileResponse{ID: 1, CompanyName: "Tech Corp", Email: "employer@example.com"},
					CreatedAt: now.Format(time.RFC3339),
					UpdatedAt: now.Format(time.RFC3339),
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
			name:        "Успешное получение вакансий по ID соискателя",
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
					WorkingHours:   18,
					SalaryFrom:     100000,
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
			name:        "Ошибка при проверке отклика",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository, es *m.MockEmployer) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]*entity.Vacancy{
						{
							ID:         1,
							EmployerID: 2,
							Title:      "Backend Developer",
						},
					}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 1, 1).
					Return(false, fmt.Errorf("ошибка базы данных"))
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
				require.Equal(t, tc.expectedErr.Error(), err.Error())
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
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

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
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
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
