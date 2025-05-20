package service

import (
	"ResuMatch/internal/config"
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
	endDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDateStr := endDate.Format("2006-01-02")

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
			name:        "Успешное создание резюме с минимальными полями",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          2,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 2).
					Return([]entity.Specialization{}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID: 2,

				ApplicantID:               1,
				Profession:                "Developer",
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{},
				AdditionalSpecializations: []string{},
				WorkExperiences:           []dto.WorkExperienceResponse{},
			},
			expectedErr: nil,
		},
		{
			name:        "Ошибка парсинга даты окончания учебы",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession:     "Developer",
				GraduationYear: "invalid-date",
			},
			mockSetup: func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository) {
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты окончания учебы: %w", fmt.Errorf("parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\"")),
			),
		},
		{
			name:        "Ошибка поиска специализации",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession:     "Developer",
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
			name:        "Ошибка создания резюме",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID: 1,
						Profession:  "Developer",
					}).
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
			name:        "Ошибка парсинга даты начала работы",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    "invalid-date",
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{ID: 1, ApplicantID: 1, Profession: "Developer"}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты начала работы: %w", fmt.Errorf("parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\"")),
			),
		},
		{
			name:        "Ошибка парсинга даты окончания работы",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDateStr,
						EndDate:      "invalid-date",
						UntilNow:     false,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{ID: 1, ApplicantID: 1, Profession: "Developer"}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты окончания работы: %w", fmt.Errorf("parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\"")),
			),
		},
		{
			name:        "Ошибка добавления опыта работы",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDateStr,
						UntilNow:     true,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{ID: 1, ApplicantID: 1, Profession: "Developer"}, nil)

				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:     1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDate,
						UntilNow:     true,
					}).
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
			name:        "Ошибка валидации резюме",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				AboutMe: "Опытный разработчик",
				// Profession is missing, assuming validation fails
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID: 1,
						AboutMe:     "Опытный разработчик",
					}).
					Return(nil, entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("ошибка валидации резюме"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка валидации резюме"),
			),
		},
		{
			name:        "Ошибка валидации опыта работы",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						// Missing EmployerName, assuming validation fails
						Position:  "Разработчик",
						StartDate: startDateStr,
						UntilNow:  true,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{ID: 1, ApplicantID: 1, Profession: "Developer"}, nil)

				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:  1,
						Position:  "Разработчик",
						StartDate: startDate,
						UntilNow:  true,
					}).
					Return(nil, entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("ошибка валидации опыта работы"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка валидации опыта работы"),
			),
		},
		{
			name:        "Успешное создание резюме с опытом работы и датой окончания",
			applicantID: 1,
			request: &dto.CreateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDateStr,
						EndDate:      endDateStr,
						UntilNow:     false,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					Create(gomock.Any(), &entity.Resume{
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          3,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:     3,
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDate,
						EndDate:      endDate,
						UntilNow:     false,
					}).
					Return(&entity.WorkExperience{
						ID:           2,
						ResumeID:     3,
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDate,
						EndDate:      endDate,
						UntilNow:     false,
						UpdatedAt:    now,
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 3).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 3).
					Return([]entity.Specialization{}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        3,
				ApplicantID:               1,
				Profession:                "Developer",
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{},
				AdditionalSpecializations: []string{},
				WorkExperiences: []dto.WorkExperienceResponse{
					{
						ID:           2,
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDateStr,
						EndDate:      endDateStr,
						UntilNow:     false,
						UpdatedAt:    now.Format(time.RFC3339),
					},
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
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockApplicantService := NewApplicantService(mockApplicantRepo, nil, nil)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo)
			var cfg = config.ResumeConfig{}
			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService, cfg)
			ctx := context.Background()

			result, err := service.Create(ctx, tc.applicantID, tc.request)

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

func TestResumeService_GetByID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	gradYear := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	gradYearStr := gradYear.Format("2006-01-02")
	startDate := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	startDateStr := startDate.Format("2006-01-02")
	endDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDateStr := endDate.Format("2006-01-02")

	testCases := []struct {
		name           string
		resumeID       int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository)
		expectedResult *dto.ResumeResponse
		expectedErr    error
	}{
		{
			name:     "Успешное получение резюме со всеми полями",
			resumeID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
						{
							ID:           2,
							ResumeID:     1,
							EmployerName: "Google",
							Position:     "Engineer",
							Duties:       "Development",
							Achievements: "Improved performance",
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
					{
						ID:           2,
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDateStr,
						EndDate:      endDateStr,
						UntilNow:     false,
						UpdatedAt:    now.Format(time.RFC3339),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:     "Успешное получение резюме с минимальными полями",
			resumeID: 2,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Resume{
						ID:          2,
						ApplicantID: 1,
						Profession:  "Developer",
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
				Profession:                "Developer",
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{},
				AdditionalSpecializations: []string{},
				WorkExperiences:           []dto.WorkExperienceResponse{},
			},
			expectedErr: nil,
		},
		{
			name:     "Резюме не найдено",
			resumeID: 999,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("резюме с id=%d не найдено", 999),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=%d не найдено", 999),
			),
		},
		{
			name:     "Ошибка получения специализации",
			resumeID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:               1,
						ApplicantID:      1,
						SpecializationID: 1,
						Profession:       "Developer",
						CreatedAt:        now,
						UpdatedAt:        now,
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
			name:     "Ошибка получения навыков",
			resumeID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
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
			name:     "Ошибка получения дополнительных специализаций",
			resumeID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении дополнительных специализаций"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении дополнительных специализаций"),
			),
		},
		{
			name:     "Ошибка получения опыта работы",
			resumeID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return([]entity.Specialization{}, nil)

				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
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

			var cfg = config.ResumeConfig{}
			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService, cfg)
			ctx := context.Background()

			result, err := service.GetByID(ctx, tc.resumeID)

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

func TestResumeService_Update(t *testing.T) {
	t.Parallel()

	now := time.Now()
	gradYear := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.UTC)
	gradYearStr := gradYear.Format("2006-01-02")
	startDate := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	startDateStr := startDate.Format("2006-01-02")
	endDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	endDateStr := endDate.Format("2006-01-02")

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
			name:        "Успешное обновление резюме со всеми полями",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
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
					{
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDateStr,
						EndDate:      endDateStr,
						UntilNow:     false,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:                     1,
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
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "SQL"}).
					Return([]int{1, 2}, nil)

				rr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
					Return(nil)

				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps"}).
					Return([]int{2}, nil)

				rr.EXPECT().
					AddSpecializations(gomock.Any(), 1, []int{2}).
					Return(nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 1).
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

				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:     1,
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDate,
						EndDate:      endDate,
						UntilNow:     false,
					}).
					Return(&entity.WorkExperience{
						ID:           2,
						ResumeID:     1,
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDate,
						EndDate:      endDate,
						UntilNow:     false,
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
					{
						ID:           2,
						EmployerName: "Google",
						Position:     "Engineer",
						Duties:       "Development",
						Achievements: "Improved performance",
						StartDate:    startDateStr,
						EndDate:      endDateStr,
						UntilNow:     false,
						UpdatedAt:    now.Format(time.RFC3339),
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Успешное обновление резюме с минимальными полями",
			resumeID:    2,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(&entity.Resume{
						ID:          2,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          2,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          2,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 2).
					Return(nil)

				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 2).
					Return(nil)

				rr.EXPECT().
					DeleteWorkExperiences(gomock.Any(), 2).
					Return(nil)

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 2).
					Return([]entity.Specialization{}, nil)
			},
			expectedResult: &dto.ResumeResponse{
				ID:                        2,
				ApplicantID:               1,
				Profession:                "Developer",
				CreatedAt:                 now.Format(time.RFC3339),
				UpdatedAt:                 now.Format(time.RFC3339),
				Skills:                    []string{},
				AdditionalSpecializations: []string{},
				WorkExperiences:           []dto.WorkExperienceResponse{},
			},
			expectedErr: nil,
		},
		{
			name:        "Резюме не найдено",
			resumeID:    999,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("резюме с id=%d не найдено", 999),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=%d не найдено", 999),
			),
		},
		{
			name:        "Запрещено: резюме не принадлежит соискателю",
			resumeID:    1,
			applicantID: 2,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrForbidden,
				fmt.Errorf("резюме с id=%d не принадлежит соискателю с id=%d", 1, 2),
			),
		},
		{
			name:        "Ошибка парсинга даты окончания учебы",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession:     "Developer",
				GraduationYear: "invalid-date",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты окончания учебы: %w", fmt.Errorf("parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\"")),
			),
		},
		{
			name:        "Ошибка валидации резюме",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				AboutMe: "Опытный разработчик",
				// Profession is missing, assuming validation fails
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						AboutMe:     "Опытный разработчик",
					}).
					Return(nil, entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("ошибка валидации резюме"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка валидации резюме"),
			),
		},
		{
			name:        "Ошибка обновления резюме",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
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
			name:        "Ошибка удаления навыков",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

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
			name:        "Ошибка поиска ID навыков",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
				Skills:     []string{"Go", "SQL"},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "SQL"}).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при поиске ID навыков"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске ID навыков"),
			),
		},
		{
			name:        "Ошибка добавления навыков",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
				Skills:     []string{"Go", "SQL"},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					FindSkillIDsByNames(gomock.Any(), []string{"Go", "SQL"}).
					Return([]int{1, 2}, nil)

				rr.EXPECT().
					AddSkills(gomock.Any(), 1, []int{1, 2}).
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
			name:        "Ошибка удаления специализаций",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

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
			name:        "Ошибка поиска ID специализаций",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession:                "Developer",
				AdditionalSpecializations: []string{"DevOps"},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps"}).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при поиске ID специализаций"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске ID специализаций"),
			),
		},
		{
			name:        "Ошибка добавления специализаций",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession:                "Developer",
				AdditionalSpecializations: []string{"DevOps"},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					FindSpecializationIDsByNames(gomock.Any(), []string{"DevOps"}).
					Return([]int{2}, nil)

				rr.EXPECT().
					AddSpecializations(gomock.Any(), 1, []int{2}).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при добавлении специализаций"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при добавлении специализаций"),
			),
		},
		{
			name:        "Ошибка удаления опыта работы",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)

				rr.EXPECT().
					DeleteSkills(gomock.Any(), 1).
					Return(nil)

				rr.EXPECT().
					DeleteSpecializations(gomock.Any(), 1).
					Return(nil)

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
			name:        "Ошибка парсинга даты начала работы",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    "invalid-date",
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
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
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты начала работы: %w", fmt.Errorf("parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\"")),
			),
		},
		{
			name:        "Ошибка парсинга даты окончания работы",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDateStr,
						EndDate:      "invalid-date",
						UntilNow:     false,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
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
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный формат даты окончания работы: %w", fmt.Errorf("parsing time \"invalid-date\" as \"2006-01-02\": cannot parse \"invalid-date\" as \"2006\"")),
			),
		},
		{
			name:        "Ошибка валидации опыта работы",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						// Missing EmployerName, assuming validation fails
						Position:  "Разработчик",
						StartDate: startDateStr,
						UntilNow:  true,
					},
				},
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
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

				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:  1,
						Position:  "Разработчик",
						StartDate: startDate,
						UntilNow:  true,
					}).
					Return(nil, entity.NewError(
						entity.ErrBadRequest,
						fmt.Errorf("ошибка валидации опыта работы"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("ошибка валидации опыта работы"),
			),
		},
		{
			name:        "Ошибка добавления опыта работы",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
				WorkExperiences: []dto.WorkExperienceDTO{
					{
						EmployerName: "Яндекс",
						Position:     "Разработчик",
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
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
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

				rr.EXPECT().
					AddWorkExperience(gomock.Any(), &entity.WorkExperience{
						ResumeID:     1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						StartDate:    startDate,
						UntilNow:     true,
					}).
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
			name:        "Ошибка получения специализации для ответа",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Specialization: "Backend разработка",
				Profession:     "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					FindSpecializationIDByName(gomock.Any(), "Backend разработка").
					Return(1, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:               1,
						ApplicantID:      1,
						SpecializationID: 1,
						Profession:       "Developer",
					}).
					Return(&entity.Resume{
						ID:               1,
						ApplicantID:      1,
						SpecializationID: 1,
						Profession:       "Developer",
						CreatedAt:        now,
						UpdatedAt:        now,
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
			name:        "Ошибка получения навыков для ответа",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
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

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
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
			name:        "Ошибка получения дополнительных специализаций для ответа",
			resumeID:    1,
			applicantID: 1,
			request: &dto.UpdateResumeRequest{
				Profession: "Developer",
			},
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Old Profession",
					}, nil)

				rr.EXPECT().
					Update(gomock.Any(), &entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
					}).
					Return(&entity.Resume{
						ID:          1,
						ApplicantID: 1,
						Profession:  "Developer",
						CreatedAt:   now,
						UpdatedAt:   now,
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

				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{}, nil)

				rr.EXPECT().
					GetSpecializationsByResumeID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении дополнительных специализаций"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении дополнительных специализаций"),
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
			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockApplicantService := NewApplicantService(mockApplicantRepo, nil, nil)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo)

			var cfg = config.ResumeConfig{}
			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService, cfg)
			ctx := context.Background()

			result, err := service.Update(ctx, tc.resumeID, tc.applicantID, tc.request)

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
func TestResumeService_Delete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		resumeID       int
		applicantID    int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository)
		expectedResult *dto.DeleteResumeResponse
		expectedErr    error
	}{
		{
			name:        "Успешное удаление резюме",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
				Message: fmt.Sprintf("Резюме с id=%d успешно удалено", 1),
			},
			expectedErr: nil,
		},
		{
			name:        "Резюме не найдено",
			resumeID:    999,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
				rr.EXPECT().
					GetByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("резюме с id=%d не найдено", 999),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("резюме с id=%d не найдено", 999),
			),
		},
		{
			name:        "Запрещено: резюме не принадлежит соискателю",
			resumeID:    1,
			applicantID: 2,
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
				fmt.Errorf("резюме с id=%d не принадлежит соискателю с id=%d", 1, 2),
			),
		},
		{
			name:        "Ошибка удаления опыта работы",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			name:        "Ошибка удаления навыков",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			name:        "Ошибка удаления специализаций",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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
			name:        "Ошибка удаления резюме",
			resumeID:    1,
			applicantID: 1,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository) {
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

			var cfg = config.ResumeConfig{}
			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService, cfg)
			ctx := context.Background()

			result, err := service.Delete(ctx, tc.resumeID, tc.applicantID)

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
func TestResumeService_GetAll(t *testing.T) {
	t.Parallel()

	now := time.Now()
	startDate := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	startDateStr := startDate.Format("2006-01-02")
	// endDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	// endDateStr := endDate.Format("2006-01-02")

	testCases := []struct {
		name           string
		limit          int
		offset         int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository, *m.MockApplicant)
		expectedResult []dto.ResumeShortResponse
		expectedErr    error
	}{
		{
			name:   "Успешное получение списка резюме",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 2,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             1,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
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
						StartDate:    startDateStr,
						UntilNow:     true,
					},
				},
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:   "Пустой список резюме",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{},
			expectedErr:    nil,
		},
		{
			name:   "Ошибка получения списка резюме",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
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
			name:   "Частичная ошибка: ошибка получения специализации",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 2,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=1 не найдена"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:   "Частичная ошибка: ошибка получения опыта работы",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 2,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience fails
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении опыта работы"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:   "Частичная ошибка: ошибка получения информации о соискателе",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 2,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Applicant fails
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("соискатель с id=1 не найден"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:   "Все резюме пропущены из-за ошибок",
			limit:  10,
			offset: 0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAll(gomock.Any(), 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:               2,
							ApplicantID:      2,
							SpecializationID: 2,
							Profession:       "Frontend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				// Resume 1: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=1 не найдена"),
					))

				// Resume 2: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=2 не найдена"),
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
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService)

			var cfg = config.ResumeConfig{}
			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService, cfg)
			ctx := context.Background()

			result, err := service.GetAll(ctx, tc.limit, tc.offset)

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
func TestResumeService_GetAllResumesByApplicantID(t *testing.T) {
	t.Parallel()

	now := time.Now()
	startDate := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	startDateStr := startDate.Format("2006-01-02")

	testCases := []struct {
		name           string
		applicantID    int
		limit          int
		offset         int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository, *m.MockApplicant)
		expectedResult []dto.ResumeApplicantShortResponse
		expectedErr    error
	}{
		{
			name:        "Успешное получение списка резюме соискателя",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "DevOps Engineer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Skills
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return([]entity.Skill{
						{Name: "Go"},
						{Name: "PostgreSQL"},
					}, nil)

				// Resume 1: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Skills (empty)
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeApplicantShortResponse{
				{
					ID:             1,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "Backend разработка",
					Profession:     "Backend Developer",
					Skills:         []string{"Go", "PostgreSQL"},
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
					WorkExperience: dto.WorkExperienceShort{
						ID:           1,
						EmployerName: "Яндекс",
						Position:     "Разработчик",
						Duties:       "Разработка сервисов",
						Achievements: "Оптимизация запросов",
						StartDate:    startDateStr,
						UntilNow:     true,
					},
				},
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "DevOps Engineer",
					Skills:         []string{},
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Пустой список резюме",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]entity.Resume{}, nil)
			},
			expectedResult: []dto.ResumeApplicantShortResponse{},
			expectedErr:    nil,
		},
		{
			name:        "Ошибка получения списка резюме",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении списка резюме для соискателя"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при получении списка резюме для соискателя"),
			),
		},
		{
			name:        "Частичная ошибка: ошибка получения специализации",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "DevOps Engineer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=1 не найдена"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Skills (empty)
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeApplicantShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "DevOps Engineer",
					Skills:         []string{},
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Частичная ошибка: ошибка получения опыта работы",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "DevOps Engineer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience fails
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении опыта работы"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Skills (empty)
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeApplicantShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "DevOps Engineer",
					Skills:         []string{},
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Частичная ошибка: ошибка получения информации о соискателе",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "DevOps Engineer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Applicant fails
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("соискатель с id=1 не найден"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Skills (empty)
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeApplicantShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "DevOps Engineer",
					Skills:         []string{},
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Частичная ошибка: ошибка получения навыков",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "DevOps Engineer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)

				// Resume 1: Skills fails
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении навыков"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Skills (empty)
				rr.EXPECT().
					GetSkillsByResumeID(gomock.Any(), 2).
					Return([]entity.Skill{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeApplicantShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "DevOps Engineer",
					Skills:         []string{},
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Все резюме пропущены из-за ошибок",
			applicantID: 1,
			limit:       10,
			offset:      0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					GetAllResumesByApplicantID(gomock.Any(), 1, 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:               2,
							ApplicantID:      1,
							SpecializationID: 2,
							Profession:       "DevOps Engineer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				// Resume 1: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=1 не найдена"),
					))

				// Resume 2: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=2 не найдена"),
					))
			},
			expectedResult: []dto.ResumeApplicantShortResponse{},
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
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService)

			var cfg = config.ResumeConfig{}
			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService, cfg)
			ctx := context.Background()

			result, err := service.GetAllResumesByApplicantID(ctx, tc.applicantID, tc.limit, tc.offset)

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
func TestResumeService_SearchResumesByProfession(t *testing.T) {
	t.Parallel()

	now := time.Now()
	startDate := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	startDateStr := startDate.Format("2006-01-02")

	type testConfig struct {
		role string
	}

	testCases := []struct {
		name           string
		config         testConfig
		userID         int
		profession     string
		limit          int
		offset         int
		mockSetup      func(*mock.MockResumeRepository, *mock.MockSkillRepository, *mock.MockSpecializationRepository, *mock.MockApplicantRepository, *m.MockApplicant)
		expectedResult []dto.ResumeShortResponse
		expectedErr    error
	}{
		{
			name:       "Успешный поиск для соискателя",
			config:     testConfig{role: "applicant"},
			userID:     1,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfessionForApplicant(gomock.Any(), 1, "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             1,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
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
						StartDate:    startDateStr,
						UntilNow:     true,
					},
				},
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Успешный поиск для работодателя",
			config:     testConfig{role: "employer"},
			userID:     0,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfession(gomock.Any(), "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 2,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             1,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
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
						StartDate:    startDateStr,
						UntilNow:     true,
					},
				},
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Пустой список резюме для соискателя",
			config:     testConfig{role: "applicant"},
			userID:     1,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfessionForApplicant(gomock.Any(), 1, "Developer", 10, 0).
					Return([]entity.Resume{}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{},
			expectedErr:    nil,
		},
		{
			name:       "Пустой список резюме для работодателя",
			config:     testConfig{role: "employer"},
			userID:     0,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfession(gomock.Any(), "Developer", 10, 0).
					Return([]entity.Resume{}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{},
			expectedErr:    nil,
		},
		{
			name:       "Ошибка поиска резюме для соискателя",
			config:     testConfig{role: "applicant"},
			userID:     1,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfessionForApplicant(gomock.Any(), 1, "Developer", 10, 0).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при поиске резюме для соискателя"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске резюме для соискателя"),
			),
		},
		{
			name:       "Ошибка поиска резюме для работодателя",
			config:     testConfig{role: "employer"},
			userID:     0,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfession(gomock.Any(), "Developer", 10, 0).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при поиске резюме"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("ошибка при поиске резюме"),
			),
		},
		{
			name:       "Частичная ошибка: ошибка получения специализации (соискатель)",
			config:     testConfig{role: "applicant"},
			userID:     1,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfessionForApplicant(gomock.Any(), 1, "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=1 не найдена"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Частичная ошибка: ошибка получения специализации (работодатель)",
			config:     testConfig{role: "employer"},
			userID:     0,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfession(gomock.Any(), "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 2,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=1 не найдена"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Частичная ошибка: ошибка получения опыта работы (соискатель)",
			config:     testConfig{role: "applicant"},
			userID:     1,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfessionForApplicant(gomock.Any(), 1, "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience fails
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении опыта работы"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Частичная ошибка: ошибка получения опыта работы (работодатель)",
			config:     testConfig{role: "employer"},
			userID:     0,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfession(gomock.Any(), "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 2,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience fails
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("ошибка при получении опыта работы"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Частичная ошибка: ошибка получения информации о соискателе (соискатель)",
			config:     testConfig{role: "applicant"},
			userID:     1,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfessionForApplicant(gomock.Any(), 1, "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 1,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Applicant fails
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("соискатель с id=1 не найден"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(&dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 1, FirstName: "Иван", LastName: "Иванов"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Частичная ошибка: ошибка получения информации о соискателе (работодатель)",
			config:     testConfig{role: "employer"},
			userID:     0,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfession(gomock.Any(), "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:          2,
							ApplicantID: 2,
							Profession:  "Frontend Developer",
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)

				// Resume 1: Specialization
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(&entity.Specialization{ID: 1, Name: "Backend разработка"}, nil)

				// Resume 1: Work Experience
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
							UntilNow:     true,
							UpdatedAt:    now,
						},
					}, nil)

				// Resume 1: Applicant fails
				as.EXPECT().
					GetUser(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("соискатель с id=1 не найден"),
					))

				// Resume 2: No specialization (SpecializationID = 0, no call to GetByID)

				// Resume 2: Work Experience (empty)
				rr.EXPECT().
					GetWorkExperienceByResumeID(gomock.Any(), 2).
					Return([]entity.WorkExperience{}, nil)

				// Resume 2: Applicant
				as.EXPECT().
					GetUser(gomock.Any(), 2).
					Return(&dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"}, nil)
			},
			expectedResult: []dto.ResumeShortResponse{
				{
					ID:             2,
					Applicant:      &dto.ApplicantProfileResponse{ID: 2, FirstName: "Петр", LastName: "Петров"},
					Specialization: "",
					Profession:     "Frontend Developer",
					CreatedAt:      now.Format(time.RFC3339),
					UpdatedAt:      now.Format(time.RFC3339),
				},
			},
			expectedErr: nil,
		},
		{
			name:       "Все резюме пропущены из-за ошибок (соискатель)",
			config:     testConfig{role: "applicant"},
			userID:     1,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfessionForApplicant(gomock.Any(), 1, "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:               2,
							ApplicantID:      1,
							SpecializationID: 2,
							Profession:       "Frontend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				// Resume 1: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=1 не найдена"),
					))

				// Resume 2: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=2 не найдена"),
					))
			},
			expectedResult: []dto.ResumeShortResponse{},
			expectedErr:    nil,
		},
		{
			name:       "Все резюме пропущены из-за ошибок (работодатель)",
			config:     testConfig{role: "employer"},
			userID:     0,
			profession: "Developer",
			limit:      10,
			offset:     0,
			mockSetup: func(rr *mock.MockResumeRepository, sr *mock.MockSkillRepository, spr *mock.MockSpecializationRepository, ar *mock.MockApplicantRepository, as *m.MockApplicant) {
				rr.EXPECT().
					SearchResumesByProfession(gomock.Any(), "Developer", 10, 0).
					Return([]entity.Resume{
						{
							ID:               1,
							ApplicantID:      1,
							SpecializationID: 1,
							Profession:       "Backend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
						{
							ID:               2,
							ApplicantID:      2,
							SpecializationID: 2,
							Profession:       "Frontend Developer",
							CreatedAt:        now,
							UpdatedAt:        now,
						},
					}, nil)

				// Resume 1: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 1).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=1 не найдена"),
					))

				// Resume 2: Specialization fails
				spr.EXPECT().
					GetByID(gomock.Any(), 2).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("специализация с id=2 не найдена"),
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
			mockApplicantService := m.NewMockApplicant(ctrl)

			tc.mockSetup(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService)

			var cfg = config.ResumeConfig{}
			service := NewResumeService(mockResumeRepo, mockSkillRepo, mockSpecRepo, mockApplicantRepo, mockApplicantService, cfg)
			ctx := context.Background()

			result, err := service.SearchResumesByProfession(ctx, tc.userID, tc.config.role, tc.profession, tc.limit, tc.offset)

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
