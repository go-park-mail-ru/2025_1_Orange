package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	m "ResuMatch/internal/usecase/mock"
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
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository)
		expectedResult *dto.ResumeResponse
		expectedErr    error
	}{
		{
			name:        "Успешное создание резюме со всеми полями",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				AboutMe:                   "Опытный разработчик",
				Specialization:            "Backend разработка",
				Profession:                "Backend Developer",
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID:            1,
						AboutMe:                "Опытный разработчик",
						SpecializationID:       1,
						Profession:             "Backend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
					}).
					Return(&entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Опытный разработчик",
						SpecializationID:       1,
						Profession:             "Backend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
						CreatedAt:              now,
						UpdatedAt:              now,
					}, nil)

				rr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "SQL"}).
					Return([]int{1, 2}, nil)

				rr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
					Return(nil)

				rr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps"}).
					Return([]int{2}, nil)

				rr.EXPECT().
					AddSpecializations(gomock.Any(), 1, []int{2}).
					Return(nil)

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

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
					}, nil)

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
				Profession:                "Backend Developer",
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
			mockSetup: func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository) {
			},
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
		{
			name:        "Успешное создание резюме без опыта работы",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				AboutMe:                "Начинающий разработчик",
				Specialization:         "Frontend разработка",
				Profession:             "Frontend Developer",
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYearStr,
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Frontend разработка").
					Return(2, nil)

				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID:            1,
						AboutMe:                "Начинающий разработчик",
						SpecializationID:       2,
						Profession:             "Frontend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
					}).
					Return(&entity.Resume{
						ID:                     2,
						ApplicantID:            1,
						AboutMe:                "Начинающий разработчик",
						SpecializationID:       2,
						Profession:             "Frontend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
						CreatedAt:              now,
						UpdatedAt:              now,
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 2).
					Return([]entity.Specialization{}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{ID: 2, Name: "Frontend разработка"}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        2,
				ApplicantID:               1,
				AboutMe:                   "Начинающий разработчик",
				Specialization:            "Frontend разработка",
				Profession:                "Frontend Developer",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYearStr,
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{},
				AdditionalSpecializations: []string{},
				WorkExperiences:           []dto.WorkExperienceResponse{},
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
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockApplicantService := NewApplicantService(mockApplicantRepo, nil, nil)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService)
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
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository)
		expectedResult *dto.ResumeResponse
		expectedErr    error
	}{
		{
			name:     "Успешное получение резюме со всеми полями",
			resumeID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Опытный разработчик",
						SpecializationID:       1,
						Profession:             "Backend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
						CreatedAt:              now,
						UpdatedAt:              now,
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{
						ID:   1,
						Name: "Backend разработка",
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
					}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return([]entity.Specialization{
						{ID: 2, Name: "DevOps"},
					}, nil)

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
				Profession:                "Backend Developer",
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
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Resume{
						ID:          2,
						ApplicantID: 1,
						AboutMe:     "Начинающий разработчик",
						Profession:  "Junior Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

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
				Profession:                "Junior Developer",
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
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 3).
					Return(&entity.Resume{
						ID:               3,
						SpecializationID: 1,
						Profession:       "Backend Developer",
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
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 4).
					Return(&entity.Resume{
						ID:         4,
						Profession: "Frontend Developer",
					}, nil)

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
		{
			name:     "Ошибка при получении дополнительных специализаций",
			resumeID: 5,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 5).
					Return(&entity.Resume{
						ID:         5,
						Profession: "Fullstack Developer",
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 5).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 5).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении специализаций"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении специализаций"),
			),
		},
		{
			name:     "Ошибка при получении опыта работы",
			resumeID: 6,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 6).
					Return(&entity.Resume{
						ID:         6,
						Profession: "DevOps Engineer",
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 6).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 6).
					Return([]entity.Specialization{}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 6).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении опыта работы"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении опыта работы"),
			),
		},
		{
			name:     "Резюме с опытом работы без даты окончания",
			resumeID: 7,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 7).
					Return(&entity.Resume{
						ID:          7,
						ApplicantID: 1,
						Profession:  "Team Lead",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 7).
					Return([]entity.Skill{
						{ID: 1, Name: "Leadership"},
					}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 7).
					Return([]entity.Specialization{}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 7).
					Return([]entity.WorkExperience{
						{
							ID:           1,
							EmployerName: "Яндекс",
							Position:     "Team Lead",
							Duties:       "Управление командой",
							Achievements: "Успешные проекты",
							StartDate:    startDate,
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        7,
				ApplicantID:               1,
				Profession:                "Team Lead",
				Skills:                    []string{"Leadership"},
				AdditionalSpecializations: []string{},
				WorkExperiences: []dto.WorkExperienceResponse{
					{
						ID:           1,
						EmployerName: "Яндекс",
						Position:     "Team Lead",
						Duties:       "Управление командой",
						Achievements: "Успешные проекты",
						StartDate:    startDate.Format("2006-01-02"),
						UntilNow:     true,
						UpdatedAt:    now.Format(time.RFC3339),
					},
				},
				CreatedAt: now.Format(time.RFC3339),
				UpdatedAt: now.Format(time.RFC3339),
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
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockApplicantService := NewApplicantService(mockApplicantRepo, nil, nil)

			tc.mockSetup(mockResumeRepo, mockSpecRepo, mockApplicantRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService)
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
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository)
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
				Profession:                "Senior Backend Developer",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYearStr,
				Skills:                    []string{"Go", "SQL", "Docker"},
				AdditionalSpecializations: []string{"DevOps", "Microservices"},
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Старший разработчик",
						Duties:       "Разработка архитектуры",
						Achievements: "Оптимизация производительности",
						StartDate:    startDateStr,
						UntilNow:     true,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Обновленный текст",
						SpecializationID:       1,
						Profession:             "Senior Backend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
					}).
					Return(&entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Обновленный текст",
						SpecializationID:       1,
						Profession:             "Senior Backend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
						CreatedAt:              now,
						UpdatedAt:              now,
					}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "SQL", "Docker"}).
					Return([]int{1, 2, 3}, nil)

				rr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2, 3}).
					Return(nil)

				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps", "Microservices"}).
					Return([]int{2, 3}, nil)

				rr.EXPECT().
					AddSpecializations(gomock.Any(), 1, []int{2, 3}).
					Return(nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:     1,
						EmployerName: "Яндекс",
						Position:     "Старший разработчик",
						Duties:       "Разработка архитектуры",
						Achievements: "Оптимизация производительности",
						StartDate:    startDate,
						UntilNow:     true,
					}).
					Return(&entity.WorkExperience{
						ID:           1,
						ResumeID:     1,
						EmployerName: "Яндекс",
						Position:     "Старший разработчик",
						Duties:       "Разработка архитектуры",
						Achievements: "Оптимизация производительности",
						StartDate:    startDate,
						UntilNow:     true,
						UpdatedAt:    now,
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{
						{ID: 1, Name: "Go"},
						{ID: 2, Name: "SQL"},
						{ID: 3, Name: "Docker"},
					}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return([]entity.Specialization{
						{ID: 2, Name: "DevOps"},
						{ID: 3, Name: "Microservices"},
					}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        1,
				ApplicantID:               1,
				AboutMe:                   "Обновленный текст",
				Specialization:            "Backend разработка",
				Profession:                "Senior Backend Developer",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYearStr,
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{"Go", "SQL", "Docker"},
				AdditionalSpecializations: []string{"DevOps", "Microservices"},
				WorkExperiences: []dto.WorkExperienceResponse{
					{
						ID:           1,
						EmployerName: "Яндекс",
						Position:     "Старший разработчик",
						Duties:       "Разработка архитектуры",
						Achievements: "Оптимизация производительности",
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
		{
			name:        "Ошибка при добавлении навыков",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Skills: []string{"Go"},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
					Return(nil)

				rr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go"}).
					Return([]int{1}, nil)

				rr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1}).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при добавлении навыков"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении навыков"),
			),
		},
		{
			name:        "Успешное обновление без опыта работы",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				AboutMe:                "Обновленный текст без опыта",
				Specialization:         "Frontend разработка",
				Profession:             "Frontend Developer",
				Education:              entity.Higher,
				EducationalInstitution: "МГУ",
				GraduationYear:         gradYearStr,
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Frontend разработка").
					Return(2, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Обновленный текст без опыта",
						SpecializationID:       2,
						Profession:             "Frontend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
					}).
					Return(&entity.Resume{
						ID:                     1,
						ApplicantID:            1,
						AboutMe:                "Обновленный текст без опыта",
						SpecializationID:       2,
						Profession:             "Frontend Developer",
						Education:              entity.Higher,
						EducationalInstitution: "МГУ",
						GraduationYear:         gradYear,
						CreatedAt:              now,
						UpdatedAt:              now,
					}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
					Return(nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{ID: 2, Name: "Frontend разработка"}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return([]entity.Specialization{}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        1,
				ApplicantID:               1,
				AboutMe:                   "Обновленный текст без опыта",
				Specialization:            "Frontend разработка",
				Profession:                "Frontend Developer",
				Education:                 entity.Higher,
				EducationalInstitution:    "МГУ",
				GraduationYear:            gradYearStr,
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{},
				AdditionalSpecializations: []string{},
				WorkExperiences:           []dto.WorkExperienceResponse{},
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
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockApplicantService := NewApplicantService(mockApplicantRepo, nil, nil)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService)
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
		mockSetup      func(*mock.MockResumeRepository, *mock.MockApplicantRepository)
		expectedResult *dto.DeleteResumeResponse
		expectedErr    error
	}{
		{
			name:        "Успешное удаление резюме",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, ar *mock.MockApplicantRepository) {
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
			mockSetup: func(rr *mock.MockResumeRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
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
			mockSetup: func(rr *mock.MockResumeRepository, ar *mock.MockApplicantRepository) {
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
						fmt.Errorf("ошибка при удалении опыта работы"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении опыта работы"),
			),
		},
		{
			name:        "Ошибка при удалении навыков",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, ar *mock.MockApplicantRepository) {
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
						fmt.Errorf("ошибка при удалении навыков"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении навыков"),
			),
		},
		{
			name:        "Ошибка при удалении специализаций",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, ar *mock.MockApplicantRepository) {
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
						fmt.Errorf("ошибка при удалении специализаций"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении специализаций"),
			),
		},
		{
			name:        "Ошибка при удалении самого резюме",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, ar *mock.MockApplicantRepository) {
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
						fmt.Errorf("ошибка при удалении резюме"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при удалении резюме"),
			),
		},
		{
			name:        "Успешное удаление резюме без связанных данных",
			resumeID:    2,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Resume{
						ID:          2,
						ApplicantID: 1,
					}, nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 2).
					Return(nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 2).
					Return(nil)

				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 2).
					Return(nil)

				rr.EXPECT().
					Delete(gomock.Any(), 2).
					Return(nil)
			},
			expectedResult: &dto.DeleteResumeResponse{
				Success: true,
				Message: "Резюме с id=2 успешно удалено",
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
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockApplicantService := NewApplicantService(mockApplicantRepo, nil, nil)

			tc.mockSetup(mockResumeRepo, mockApplicantRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService)
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
		limit          int
		offset         int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository)
		expectedResult []dto.ResumeShortResponse
		expectedErr    error
	}{
		{
			name:   "Успешное получение списка резюме",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      100,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:               2,
							ApplicantID:      101,
							SpecializationID: 2,
							Profession:       "DevOps Engineer",
							CreatedAt:        now.Add(-24 * time.Hour),
							UpdatedAt:        now.Add(-24 * time.Hour),
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

				ar.EXPECT().
					GetApplicantByID(gomock.Any(), 100).
					Return(&entity.Applicant{
						ID:        100,
						FirstName: "Иван",
						LastName:  "Иванов",
						Email:     "ivan@example.com",
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{
						ID:   2,
						Name: "DevOps",
					}, nil)

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

				ar.EXPECT().
					GetApplicantByID(gomock.Any(), 101).
					Return(&entity.Applicant{
						ID:        101,
						FirstName: "Петр",
						LastName:  "Петров",
						Email:     "petr@example.com",
					}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             1,
					Applicant:      &dto.ApplicantProfileResponse{ID: 100, FirstName: "Иван", LastName: "Иванов", Email: "ivan@example.com"},
					Specialization: "Backend разработка",
					Profession:     "Backend Developer",
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
					Applicant:      &dto.ApplicantProfileResponse{ID: 101, FirstName: "Петр", LastName: "Петров", Email: "petr@example.com"},
					Specialization: "DevOps",
					Profession:     "DevOps Engineer",
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
			name:   "Ошибка при получении списка резюме",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
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
			name:   "Ошибка при получении специализации (пропускаем резюме)",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      100,
							SpecializationID: 1,
							Profession:       "Backend Developer",
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
			name:   "Ошибка при получении опыта работы (пропускаем резюме)",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      100,
							SpecializationID: 1,
							Profession:       "Backend Developer",
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
			name:   "Резюме без специализации и опыта работы",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:          1,
							ApplicantID: 100,
							Profession:  "Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return([]entity.WorkExperience{}, nil)

				ar.EXPECT().
					GetApplicantByID(gomock.Any(), 100).
					Return(&entity.Applicant{
						ID:        100,
						FirstName: "Иван",
						LastName:  "Иванов",
						Email:     "ivan@example.com",
					}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:         1,
					Applicant:  &dto.ApplicantProfileResponse{ID: 100, FirstName: "Иван", LastName: "Иванов", Email: "ivan@example.com"},
					Profession: "Developer",
					CreatedAt:  now.Format(time.RFC3339),
					UpdatedAt:  now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка при получении информации о соискателе (пропускаем резюме)",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:          1,
							ApplicantID: 100,
							Profession:  "Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return([]entity.WorkExperience{}, nil)

				ar.EXPECT().
					GetApplicantByID(gomock.Any(), 100).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении информации о соискателе"),
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
			mockSkillRepo := mock.NewMockSkillRepository(ctrl)
			mockSpecRepo := mock.NewMockSpecializationRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockApplicantService := NewApplicantService(mockApplicantRepo, nil, nil)

			tc.mockSetup(mockResumeRepo, mockSpecRepo, mockApplicantRepo)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService)
			ctx := context.Background()

			result, err := service.GetAll(ctx, tc.limit, tc.offset)

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
	testCases := []struct {
		name           string
		applicantID    int
		limit          int
		offset         int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSpecializationRepository, *m.MockApplicant)
		expectedResult []dto.ResumeShortResponse
		expectedErr    error
	}{
		{
			name:        "Успешное получение вакансий с полной информацией",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, am *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							Title:            "Backend Developer",
							EmployerID:       100,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
							City:             "Москва",
						},
						{
							ID:               2,
							Title:            "Frontend Developer",
							EmployerID:       101,
							SpecializationID: 2,
							Profession:       "Frontend Developer",
							CreatedAt:        now.Add(-time.Hour),
							UpdatedAt:        now.Add(-time.Hour),
							City:             "Санкт-Петербург",
						},
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

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

				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Specialization{ID: 2, Name: "Frontend разработка"}, nil)

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

				am.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{
						ID:        1,
						FirstName: "Иван",
						LastName:  "Иванов",
						Email:     "ivan@example.com",
					}, nil).
					Times(2)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:             1,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов", Email: "ivan@example.com"},
					Specialization: "Backend разработка",
					Profession:     "Backend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					City:           "Москва",
					Responded:      true,
				},
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов", Email: "ivan@example.com"},
					Specialization: "Frontend разработка",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Add(-time.Hour).Format(time.RFC3339),
					UpdatedAt:      now.Add(-time.Hour).Format(time.RFC3339),
					City:           "Санкт-Петербург",
					Responded:      false,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Вакансия без специализации",
			applicantID: 2,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, am *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 2, 10, 0).
					Return([]entity.Resume{
						{
							ID:          3,
							ApplicantID: 2,
							Profession:  "Fullstack Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 3).
					Return([]entity.WorkExperience{}, nil)

				am.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{
						ID:        2,
						FirstName: "Петр",
						LastName:  "Петров",
						Email:     "petr@example.com",
					}, nil)
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:         3,
					Applicant:  &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров", Email: "petr@example.com"},
					Profession: "Fullstack Developer",
					CreatedAt:  now.Format(time.RFC3339),
					UpdatedAt:  now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка при получении вакансий",
			applicantID: 3,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, am *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 3, 10, 0).
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
			name:        "Ошибка при получении специализации (пропускаем вакансию)",
			applicantID: 4,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, am *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 4, 10, 0).
					Return([]entity.Resume{
						{
							ID:               4,
							SpecializationID: 1,
							Profession:       "DevOps Engineer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				sr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении специализации"),
					))

				am.EXPECT().
					GetUser(gomock.Any(), 4).
					Return(&dto.ApplicantProfileResponse{
						ID:        4,
						FirstName: "Сергей",
						LastName:  "Сергеев",
						Email:     "sergey@example.com",
					}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:         4,
					Applicant:  &dto.ApplicantProfileResponse{ID: 4, FirstName: "Сергей", LastName: "Сергеев", Email: "sergey@example.com"},
					Profession: "DevOps Engineer",
					CreatedAt:  now.Format(time.RFC3339),
					UpdatedAt:  now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка при проверке отклика",
			applicantID: 5,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, am *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 5, 10, 0).
					Return([]entity.Resume{
						{
							ID:               5,
							ApplicantID:      5,
							SpecializationID: 1,
							Profession:       "Data Scientist",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Data Science"}, nil)

				vr.EXPECT().
					ResponseExists(gomock.Any(), 5, 5).
					Return(false, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при проверке отклика"),
					))

				am.EXPECT().
					GetUser(gomock.Any(), 5).
					Return(&dto.ApplicantProfileResponse{
						ID:        5,
						FirstName: "Алексей",
						LastName:  "Алексеев",
						Email:     "alex@example.com",
					}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             5,
					Applicant:      &dto.ApplicantProfileResponse{ID: 5, FirstName: "Алексей", LastName: "Алексеев", Email: "alex@example.com"},
					Specialization: "Data Science",
					Profession:     "Data Scientist",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Пустой список резюме",
			applicantID: 6,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, spr *mock.MockSpecializationRepository, am *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 6, 10, 0).
					Return([]entity.Resume{}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при проверке отклика"),
			),
		},
		{
			name:        "Неавторизованный пользователь (responded=false)",
			applicantID: 0,
			mockSetup: func(vr *mock.MockVacancyRepository, sr *mock.MockSpecializationRepository) {
				vr.EXPECT().
					GetVacanciesByApplicantID(gomock.Any(), 0).
					Return([]entity.Vacancy{
						{
							ID: 6,
						},
					}, nil)

				// ResponseExists не должен вызываться для applicantID=0
			},
			expectedResult: []dto.VacancyShortResponse{
				{
					ID:        6,
					Responded: false,
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
			mockApplicant := m.NewMockApplicant(ctrl)
			mockSkillRepo := mock.NewMockSkillRepository(ctrl)
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)

			tc.mockSetup(mockResumeRepo, mockSpecRepo, mockApplicant)

			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicant)
			ctx := context.Background()

			result, err := service.GetAllResumesByApplicantID(ctx, tc.applicantID, tc.limit, tc.offset)

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
