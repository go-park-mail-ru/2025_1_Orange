package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	"context"

	// "errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResumeService_Create(t *testing.T) {
	t.Parallel()

	now := time.Now()
	gradYear := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	gradYearStr := gradYear.Format("2006-01-02")
	startDate := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	startDateStr := startDate.Format("2006-01-02")

	testCases := []struct {
		name           string
		applicantID    int
		request        *dto.CreateResumeRequest
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository)
		expectedResult *dto.ResumeResponse
		expectedErr    error
	}{
		{
			name:        "Успешное создание резюме со всеми полями",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				AboutMe:                   "Опытный разработчик",
				Specialization:            "Backend разработка",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYearStr,
				Skills:                    []string{"Go", "SQL"},
				AdditionalSpecializations: []string{"DevOps"},
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						Duties:       "Разработка сервисов",
						Achievements: "Оптимизация запросов",
						StartDate:    startDateStr,
						UntilNow:     true,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				// Поиск специализации
				rr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				// Создание резюме
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID:            1,
						AboutMe:                "Опытный разработчик",
						SpecializationID:       1,
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
					}).
					Return(&entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Опытный разработчик",
						SpecializationID:       1,
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
						CreatedAt:              now,
						UpdatedAt:              now,
					}, nil)

				// Поиск ID навыков
				rr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "SQL"}).
					Return([]int{1, 2}, nil)

				// Добавление навыков
				rr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
					Return(nil)

				// Поиск ID дополнительных специализаций
				rr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps"}).
					Return([]int{2}, nil)

				// Добавление специализаций
				rr.EXPECT().
					AddSpecializations(gomock.Any(), 1, []int{2}).
					Return(nil)

				// Добавление опыта работы
				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:     1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						Duties:       "Разработка сервисов",
						Achievements: "Оптимизация запросов",
						StartDate:    startDate,
						UntilNow:     true,
					}).
					Return(&entity.WorkExperience{
						ID:           1,
						ResumeID:     1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						Duties:       "Разработка сервисов",
						Achievements: "Оптимизация запросов",
						StartDate:    startDate,
						UntilNow:     true,
						UpdatedAt:    now,
					}, nil)

				// Получение названия специализации
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Получение навыков
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
					}, nil)

				// Получение дополнительных специализаций
				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return([]entity.Specialization{
						{ID: 2, Name: "DevOps"},
					}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        1,
				ApplicantID:               1,
				AboutMe:                   "Опытный разработчик",
				Specialization:            "Backend разработка",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYearStr,
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{"Go", "SQL"},
				AdditionalSpecializations: []string{"DevOps"},
				WorkExperiences: []dto.WorkExperienceResponse{
					{
						ID:           1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						Duties:       "Разработка сервисов",
						Achievements: "Оптимизация запросов",
						StartDate:    startDateStr,
						UntilNow:     true,
						UpdatedAt:    now.Format(time.RFC3339),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка при парсинге даты окончания",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				GraduationYear: "invalid-date",
			},
			mockSetup:      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository) {},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты окончания учебы: parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\""),
			),
		},
		{
			name:        "Ошибка при поиске специализации",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Specialization: "Backend разработка",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
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
			name:        "Ошибка при создании резюме",
			applicantID: 1,
			request:     &dto.CreateResumeRequest{},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при создании резюме"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при создании резюме"),
			),
		},
		{
			name:        "Ошибка при парсинге даты начала работы",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						StartDate: "invalid-date",
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Resume{ID: 1}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты начала работы: parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\""),
			),
		},
		{
			name:        "Ошибка при добавлении опыта работы",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						StartDate: startDateStr,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&entity.Resume{ID: 1}, nil)

				rr.EXPECT().
					AddWorkExperience(gomock.Any(), gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при добавлении опыта работы"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении опыта работы"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockSkillRepo := mock.NewMockSkillRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.Create(ctx, tc.applicantID, tc.request)

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

func TestResumeService_GetByID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	gradYear := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2021, time.December, 31, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		resumeID       int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSpecializationRepository)
		expectedResult *dto.ResumeResponse
		expectedErr    error
	}{
		{
			name:     "Успешное получение резюме со всеми полями",
			resumeID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				// Получение резюме
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Опытный разработчик",
						SpecializationID:       1,
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
						CreatedAt:              now,
						UpdatedAt:              now,
					}, nil)

				// Получение специализации
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				// Получение навыков
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
					}, nil)

				// Получение дополнительных специализаций
				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return([]entity.Specialization{
						{ID: 2, Name: "DevOps"},
					}, nil)

				// Получение опыта работы
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return([]entity.WorkExperience{
						{
							ID:           1,
							EmployerName: "Яндекс",
							Position:     "Разработчик",
							Duties:       "Разработка сервисов",
							Achievements: "Оптимизация запросов",
							StartDate:    startDate,
							EndDate:      endDate,
							UntilNow:     false,
							UpdatedAt:    now,
						},
					}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        1,
				ApplicantID:               1,
				AboutMe:                   "Опытный разработчик",
				Specialization:            "Backend разработка",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYear.Format("2006-01-02"),
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{"Go", "SQL"},
				AdditionalSpecializations: []string{"DevOps"},
				WorkExperiences: []dto.WorkExperienceResponse{
					{
						ID:           1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						Duties:       "Разработка сервисов",
						Achievements: "Оптимизация запросов",
						StartDate:    startDate.Format("2006-01-02"),
						EndDate:      endDate.Format("2006-01-02"),
						UntilNow:     false,
						UpdatedAt:    now.Format(time.RFC3339),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:     "Резюме без специализации и опыта работы",
			resumeID: 2,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Resume{
						ID:          2,
						ApplicantID: 1,
						AboutMe:     "Начинающий разработчик",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				// Специализация не вызывается, так как SpecializationID = 0

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 2).
					Return([]entity.Specialization{}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        2,
				ApplicantID:               1,
				AboutMe:                   "Начинающий разработчик",
				Skills:                    []string{},
				AdditionalSpecializations: []string{},
				WorkExperiences:           []dto.WorkExperienceResponse{},
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
			},
			expectedErr: nil,
		},
		{
			name:     "Резюме не найдено",
			resumeID: 999,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("резюме с id=999 не найдено"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=999 не найдено"),
			),
		},
		{
			name:     "Ошибка при получении специализации",
			resumeID: 3,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 3).
					Return(&entity.Resume{
						ID:               3,
						SpecializationID: 1,
					}, nil)

				spr.EXPECT().
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
		{
			name:     "Ошибка при получении навыков",
			resumeID: 4,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 4).
					Return(&entity.Resume{ID: 4}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 4).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении навыков"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении навыков"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockSkillRepo := mock.NewMockSkillRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockResumeRepo, mockSpecRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.GetByID(ctx, tc.resumeID)

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

func TestResumeService_Update(t *testing.T) {
	t.Parallel()

	now := time.Now()
	gradYear := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	gradYearStr := gradYear.Format("2006-01-02")
	startDate := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	startDateStr := startDate.Format("2006-01-02")

	testCases := []struct {
		name           string
		resumeID       int
		applicantID    int
		request        *dto.UpdateResumeRequest
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository)
		expectedResult *dto.ResumeResponse
		expectedErr    error
	}{
		{
			name:        "Успешное обновление резюме",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				AboutMe:                   "Обновленный текст",
				Specialization:            "Backend разработка",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYearStr,
				Skills:                    []string{"Go", "SQL"},
				AdditionalSpecializations: []string{"DevOps"},
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDateStr,
						UntilNow:     true,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				// Проверка существования резюме
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				// Поиск специализации
				rr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				// Обновление резюме
				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Обновленный текст",
						SpecializationID:       1,
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
					}).
					Return(&entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Обновленный текст",
						SpecializationID:       1,
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
						CreatedAt:              now,
						UpdatedAt:              now,
					}, nil)

				// Удаление старых навыков
				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				// Поиск новых навыков
				rr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "SQL"}).
					Return([]int{1, 2}, nil)

				// Добавление новых навыков
				rr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
					Return(nil)

				// Удаление старых специализаций
				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)

				// Поиск новых специализаций
				rr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps"}).
					Return([]int{2}, nil)

				// Добавление новых специализаций
				rr.EXPECT().
					AddSpecializations(gomock.Any(), 1, []int{2}).
					Return(nil)

				// Удаление старого опыта работы
				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
					Return(nil)

				// Добавление нового опыта работы
				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:     1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDate,
						UntilNow:     true,
					}).
					Return(&entity.WorkExperience{
						ID:           1,
						ResumeID:     1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDate,
						UntilNow:     true,
						UpdatedAt:    now,
					}, nil)

				// Получение названия специализации
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Получение навыков
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
					}, nil)

				// Получение дополнительных специализаций
				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return([]entity.Specialization{
						{ID: 2, Name: "DevOps"},
					}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        1,
				ApplicantID:               1,
				AboutMe:                   "Обновленный текст",
				Specialization:            "Backend разработка",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYearStr,
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{"Go", "SQL"},
				AdditionalSpecializations: []string{"DevOps"},
				WorkExperiences: []dto.WorkExperienceResponse{
					{
						ID:           1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDateStr,
						UntilNow:     true,
						UpdatedAt:    now.Format(time.RFC3339),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Резюме не найдено",
			resumeID:    999,
			applicantID: 1,
			request:     &dto.UpdateResumeRequest{},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("резюме с id=999 не найдено"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=999 не найдено"),
			),
		},
		{
			name:        "Резюме не принадлежит пользователю",
			resumeID:    1,
			applicantID: 2,
			request:     &dto.UpdateResumeRequest{},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1, // Другой applicantID
					}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrForbidden,
				fmt.Errorf("резюме с id=1 не принадлежит соискателю с id=2"),
			),
		},
		{
			name:        "Ошибка при парсинге даты окончания",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				GraduationYear: "invalid-date",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
					}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты окончания учебы: parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\""),
			),
		},
		{
			name:        "Ошибка при обновлении резюме",
			resumeID:    1,
			applicantID: 1,
			request:     &dto.UpdateResumeRequest{},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при обновлении резюме"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при обновлении резюме"),
			),
		},
		{
			name:        "Ошибка при удалении навыков",
			resumeID:    1,
			applicantID: 1,
			request:     &dto.UpdateResumeRequest{},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(&entity.Resume{ID: 1}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при удалении навыков"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении навыков"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockSkillRepo := mock.NewMockSkillRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.Update(ctx, tc.resumeID, tc.applicantID, tc.request)

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

func TestResumeService_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		resumeID       int
		applicantID    int
		mockSetup      func(*mock.MockResumeRepository)
		expectedResult *dto.DeleteResumeResponse
		expectedErr    error
	}{
		{
			name:        "Успешное удаление резюме",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository) {
				// Проверка существования резюме
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
					}, nil)

				// Удаление связанных данных
				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
					Return(nil)
				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)
				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)

				// Удаление самого резюме
				rr.EXPECT().
					Delete(gomock.Any(), 1).
					Return(nil)
			},
			expectedResult: &dto.DeleteResumeResponse{
				Success: true,
				Message: "Резюме с id=1 успешно удалено",
			},
			expectedErr: nil,
		},
		{
			name:        "Резюме не найдено",
			resumeID:    999,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("резюме с id=999 не найдено"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=999 не найдено"),
			),
		},
		{
			name:        "Резюме не принадлежит соискателю",
			resumeID:    1,
			applicantID: 2,
			mockSetup: func(rr *mock.MockResumeRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1, // Другой applicantID
					}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrForbidden,
				fmt.Errorf("резюме с id=1 не принадлежит соискателю с id=2"),
			),
		},
		{
			name:        "Ошибка при удалении опыта работы",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
					}, nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("database error"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("database error"),
			),
		},
		{
			name:        "Ошибка при удалении навыков",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
					}, nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
					Return(nil)
				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("skills delete error"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("skills delete error"),
			),
		},
		{
			name:        "Ошибка при удалении специализаций",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
					}, nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
					Return(nil)
				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)
				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("specializations delete error"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("specializations delete error"),
			),
		},
		{
			name:        "Ошибка при удалении самого резюме",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
					}, nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
					Return(nil)
				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)
				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)
				rr.EXPECT().
					Delete(gomock.Any(), 1).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("resume delete error"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("resume delete error"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockSkillRepo := mock.NewMockSkillRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockResumeRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.Delete(ctx, tc.resumeID, tc.applicantID)

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

func TestResumeService_GetAll(t *testing.T) {
	t.Parallel()

	now := time.Now()
	startDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSpecializationRepository)
		expectedResult []dto.ResumeShortResponse
		expectedErr    error
	}{
		{
			name: "Успешное получение списка резюме",
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				// Получение списка резюме
				rr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      100,
							SpecializationID: 1,
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:               2,
							ApplicantID:      101,
							SpecializationID: 2,
							CreatedAt:        now.Add(-24 * time.Hour),
							UpdatedAt:        now.Add(-24 * time.Hour),
						},
					}, nil)

				// Получение специализации для первого резюме
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				// Получение опыта работы для первого резюме
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return([]entity.WorkExperience{
						{
							ID:           1,
							ResumeID:     1,
							EmployerName: "Яндекс",
							Position:     "Разработчик",
							Duties:       "Разработка сервисов",
							Achievements: "Оптимизация запросов",
							StartDate:    startDate,
							EndDate:      endDate,
							UntilNow:     false,
						},
					}, nil)

				// Получение специализации для второго резюме
				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{
						ID:   2,
						Name: "DevOps",
					}, nil)

				// Получение опыта работы для второго резюме
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{
						{
							ID:           2,
							ResumeID:     2,
							EmployerName: "Google",
							Position:     "Инженер",
							Duties:       "Поддержка инфраструктуры",
							Achievements: "Автоматизация деплоя",
							StartDate:    startDate,
							UntilNow:     true,
						},
					}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             1,
					ApplicantID:    100,
					Specialization: "Backend разработка",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					WorkExperience: dto.WorkExperienceShort{
						ID:           1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						Duties:       "Разработка сервисов",
						Achievements: "Оптимизация запросов",
						StartDate:    startDate.Format("2006-01-02"),
						EndDate:      endDate.Format("2006-01-02"),
						UntilNow:     false,
					},
				},
				{
					ID:             2,
					ApplicantID:    101,
					Specialization: "DevOps",
					CreatedAt:      now.Add(-24 * time.Hour).Format(time.RFC3339),
					UpdatedAt:      now.Add(-24 * time.Hour).Format(time.RFC3339),
					WorkExperience: dto.WorkExperienceShort{
						ID:           2,
						EmployerName: "Google",
						Position:     "Инженер",
						Duties:       "Поддержка инфраструктуры",
						Achievements: "Автоматизация деплоя",
						StartDate:    startDate.Format("2006-01-02"),
						UntilNow:     true,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Ошибка при получении списка резюме",
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetAll(gomock.Any()).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении списка резюме"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка резюме"),
			),
		},
		{
			name: "Ошибка при получении специализации (пропускаем резюме)",
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      100,
							SpecializationID: 1,
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении специализации"),
					))
			},
			expectedResult: []dto.ResumeShortResponse{},
			expectedErr:    nil,
		},
		{
			name: "Ошибка при получении опыта работы (пропускаем резюме)",
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      100,
							SpecializationID: 1,
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении опыта работы"),
					))
			},
			expectedResult: []dto.ResumeShortResponse{},
			expectedErr:    nil,
		},
		{
			name: "Резюме без специализации и опыта работы",
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetAll(gomock.Any()).
					Return([]entity.Resume{
						{
							ID:          1,
							ApplicantID: 100,
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return([]entity.WorkExperience{}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:          1,
					ApplicantID: 100,
					CreatedAt:   now.Format(time.RFC3339),
					UpdatedAt:   now.Format(time.RFC3339),
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

			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockSkillRepo := mock.NewMockSkillRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)

			tc.mockSetup(mockResumeRepo, mockSpecRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.GetAll(ctx)

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

func TestResumeService_GetAllResumesByApplicantID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	startDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, time.December, 31, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		applicantID    int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSpecializationRepository)
		expectedResult []dto.ResumeShortResponse
		expectedErr    error
	}{
		{
			name:        "Успешное получение списка резюме с опытом работы",
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				// Получение резюме
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:               2,
							ApplicantID:      1,
							SpecializationID: 2,
							CreatedAt:        now.Add(-time.Hour),
							UpdatedAt:        now.Add(-time.Hour),
						},
					}, nil)

				// Получение специализаций
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)
				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{ID: 2, Name: "Frontend разработка"}, nil)

				// Получение опыта работы
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return([]entity.WorkExperience{
						{
							ID:           1,
							ResumeID:     1,
							EmployerName: "Яндекс",
							Position:     "Разработчик",
							Duties:       "Разработка сервисов",
							Achievements: "Оптимизация запросов",
							StartDate:    startDate,
							EndDate:      endDate,
							UntilNow:     false,
						},
					}, nil)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{
						{
							ID:           2,
							ResumeID:     2,
							EmployerName: "Google",
							Position:     "Инженер",
							Duties:       "Разработка интерфейсов",
							Achievements: "Улучшение UX",
							StartDate:    startDate,
							UntilNow:     true,
						},
					}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             1,
					ApplicantID:    1,
					Specialization: "Backend разработка",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					WorkExperience: dto.WorkExperienceShort{
						ID:           1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						Duties:       "Разработка сервисов",
						Achievements: "Оптимизация запросов",
						StartDate:    startDate.Format("2006-01-02"),
						EndDate:      endDate.Format("2006-01-02"),
						UntilNow:     false,
					},
				},
				{
					ID:             2,
					ApplicantID:    1,
					Specialization: "Frontend разработка",
					CreatedAt:      now.Add(-time.Hour).Format(time.RFC3339),
					UpdatedAt:      now.Add(-time.Hour).Format(time.RFC3339),
					WorkExperience: dto.WorkExperienceShort{
						ID:           2,
						EmployerName: "Google",
						Position:     "Инженер",
						Duties:       "Разработка интерфейсов",
						Achievements: "Улучшение UX",
						StartDate:    startDate.Format("2006-01-02"),
						UntilNow:     true,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Резюме без специализации и опыта работы",
			applicantID: 2,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 2).
					Return([]entity.Resume{
						{
							ID:          3,
							ApplicantID: 2,
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 3).
					Return([]entity.WorkExperience{}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:          3,
					ApplicantID: 2,
					CreatedAt:   now.Format(time.RFC3339),
					UpdatedAt:   now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка при получении списка резюме",
			applicantID: 3,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 3).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении списка резюме"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка резюме"),
			),
		},
		{
			name:        "Ошибка при получении специализации (пропускаем резюме)",
			applicantID: 4,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 4).
					Return([]entity.Resume{
						{
							ID:               4,
							ApplicantID:      4,
							SpecializationID: 1,
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении специализации"),
					))

				// Опыт работы не должен запрашиваться, так как есть continue после ошибки специализации
			},
			expectedResult: []dto.ResumeShortResponse{}, // Ожидаем пустой список, так как резюме пропущено
			expectedErr:    nil,
		},
		{
			name:        "Ошибка при получении опыта работы (пропускаем опыт)",
			applicantID: 5,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 5).
					Return([]entity.Resume{
						{
							ID:               5,
							ApplicantID:      5,
							SpecializationID: 1,
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 5).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении опыта работы"),
					))
			},
			expectedResult: []dto.ResumeShortResponse{},
			expectedErr:    nil,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockResumeRepo := mock.NewMockResumeRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			mockSkillRepo := mock.NewMockSkillRepository(ctrl)

			tc.mockSetup(mockResumeRepo, mockSpecRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo)
			ctx := context.Background()

			result, err := service.GetAllResumesByApplicantID(ctx, tc.applicantID)

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
