package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	mockUC "ResuMatch/internal/usecase/mock"
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"strings"
	"testing"
	"time"
)

func TestApplicantService_Register(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       *dto.ApplicantRegister
		mockSetup   func(*mock.MockApplicantRepository)
		expectedID  int
		expectedErr error
	}{
		{
			name: "Успешная регистрация",
			input: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "valid@email.com",
					Password: "validpasswordI!",
				},
				FirstName: "Test",
				LastName:  "User",
			},
			mockSetup: func(m *mock.MockApplicantRepository) {
				m.EXPECT().
					CreateApplicant(
						gomock.Any(),
						"valid@email.com",
						"Test",
						"User",
						gomock.Any(),
						gomock.Any(),
					).
					Return(&entity.Applicant{ID: 1}, nil)
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name: "Неправильный формат почты",
			input: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "invalid_email.com",
					Password: "validpasswordI!",
				},
				FirstName: "Test",
				LastName:  "User",
			},
			mockSetup:  func(m *mock.MockApplicantRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("невалидная почта"),
			),
		},
		{
			name: "Неправильный формат пароля",
			input: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "valid@email.com",
					Password: "short",
				},
				FirstName: "Test",
				LastName:  "User",
			},
			mockSetup:  func(m *mock.MockApplicantRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("пароль должен содержать не менее 8 символов"),
			),
		},
		{
			name: "Слишком длинное имя",
			input: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "valid@email.com",
					Password: "validPasswordI!",
				},
				FirstName: strings.Repeat("a", 31),
				LastName:  "User",
			},
			mockSetup:  func(m *mock.MockApplicantRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных: first_name: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa does not validate as runelength(2|30)"),
			),
		},
		{
			name: "Слишком длинная фамилия",
			input: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "valid@email.com",
					Password: "validpasswordI!",
				},
				FirstName: "Test",
				LastName:  strings.Repeat("a", 31),
			},
			mockSetup:  func(m *mock.MockApplicantRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных: last_name: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa does not validate as runelength(2|30)"),
			),
		},
		{
			name: "Ошибка регистрации пользователя",
			input: &dto.ApplicantRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "valid@email.com",
					Password: "validpasswordI!",
				},
				FirstName: "Test",
				LastName:  "User",
			},
			mockSetup: func(m *mock.MockApplicantRepository) {
				m.EXPECT().
					CreateApplicant(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						errors.New("неизвестная ошибка при создании соискателя err=pq: test pq error"),
					),
					)
			},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				errors.New("неизвестная ошибка при создании соискателя err=pq: test pq error"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockApplicantRepository(ctrl)
			applicantService := NewApplicantService(mockRepo, nil, nil)

			tc.mockSetup(mockRepo)

			id, err := applicantService.Register(context.Background(), tc.input)

			require.Equal(t, tc.expectedID, id)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var serviceErr entity.Error
				require.ErrorAs(t, err, &serviceErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApplicantService_Login(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       *dto.Login
		mockSetup   func(*mock.MockApplicantRepository)
		expectedID  int
		expectedErr error
	}{
		{
			name: "Успешная авторизация",
			input: &dto.Login{
				Email:    "valid@email.com",
				Password: "validpasswordI!",
			},
			mockSetup: func(m *mock.MockApplicantRepository) {
				salt, hash, _ := entity.HashPassword("validpasswordI!")
				m.EXPECT().
					GetApplicantByEmail(gomock.Any(), "valid@email.com").
					Return(&entity.Applicant{
						ID:           1,
						PasswordHash: hash,
						PasswordSalt: salt,
					}, nil)
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name: "Неправильный формат почты",
			input: &dto.Login{
				Email:    "invalid_email.com",
				Password: "validpasswordI!",
			},
			mockSetup:  func(m *mock.MockApplicantRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("невалидная почта"),
			),
		},
		{
			name: "Неправильный формат пароля",
			input: &dto.Login{
				Email:    "valid@email.com",
				Password: "short",
			},
			mockSetup:  func(m *mock.MockApplicantRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("пароль должен содержать не менее 8 символов"),
			),
		},
		{
			name: "Пользователь не найден",
			input: &dto.Login{
				Email:    "notfound@email.com",
				Password: "anypassword",
			},
			mockSetup: func(m *mock.MockApplicantRepository) {
				m.EXPECT().
					GetApplicantByEmail(gomock.Any(), "notfound@email.com").
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("соискатель с email=notfound@email.com не найден"),
					),
					)
			},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("соискатель с email=notfound@email.com не найден"),
			),
		},
		{
			name: "Неверный пароль",
			input: &dto.Login{
				Email:    "valid@email.com",
				Password: "wrongPassword123!",
			},
			mockSetup: func(m *mock.MockApplicantRepository) {
				salt, hash, _ := entity.HashPassword("correctPassword123!")
				m.EXPECT().
					GetApplicantByEmail(gomock.Any(), "valid@email.com").
					Return(&entity.Applicant{
						ID:           1,
						PasswordHash: hash,
						PasswordSalt: salt,
					}, nil)
			},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrForbidden,
				fmt.Errorf("неверный пароль"),
			),
		},
		{
			name: "Ошибка получения пользователя",
			input: &dto.Login{
				Email:    "valid@email.com",
				Password: "validpasswordI!",
			},
			mockSetup: func(m *mock.MockApplicantRepository) {
				m.EXPECT().
					GetApplicantByEmail(gomock.Any(), "valid@email.com").
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось найти соискателя с email=valid@email.com"),
					),
					)
			},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось найти соискателя с email=valid@email.com"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockApplicantRepository(ctrl)
			applicantService := NewApplicantService(mockRepo, nil, nil)

			tc.mockSetup(mockRepo)

			id, err := applicantService.Login(context.Background(), tc.input)

			require.Equal(t, tc.expectedID, id)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var serviceErr entity.Error
				require.ErrorAs(t, err, &serviceErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApplicantService_UpdateProfile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		userID      int
		input       *dto.ApplicantProfileUpdate
		mockSetup   func(*mock.MockApplicantRepository, *mock.MockCityRepository)
		expectedErr error
	}{
		{
			name:   "Успешное обновление всех полей",
			userID: 1,
			input: &dto.ApplicantProfileUpdate{
				FirstName:  "НовоеИмя",
				LastName:   "НоваяФамилия",
				MiddleName: "НовоеОтчество",
				BirthDate:  time.Date(1990, 4, 2, 0, 0, 0, 0, time.UTC),
				Sex:        "M",
				Status:     "actively_searching",
				Quote:      "Один за всех и все за одного",
				Vk:         "https://vk.com/updated_profile",
				Telegram:   "https://t.me/updated_profile",
				Facebook:   "https://facebook.com/updated_profile",
				City:       "Москва",
			},
			mockSetup: func(applicantRepo *mock.MockApplicantRepository, cityRepo *mock.MockCityRepository) {
				cityRepo.EXPECT().
					GetCityByName(gomock.Any(), "Москва").
					Return(&entity.City{ID: 1}, nil)

				applicantRepo.EXPECT().
					UpdateApplicant(gomock.Any(), 1, map[string]interface{}{
						"first_name":  "НовоеИмя",
						"last_name":   "НоваяФамилия",
						"middle_name": "НовоеОтчество",
						"birth_date":  time.Date(1990, 4, 2, 0, 0, 0, 0, time.UTC),
						"sex":         "M",
						"status":      "actively_searching",
						"quote":       "Один за всех и все за одного",
						"vk":          "https://vk.com/updated_profile",
						"telegram":    "https://t.me/updated_profile",
						"facebook":    "https://facebook.com/updated_profile",
						"city_id":     1,
					}).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Успешное обновление города",
			userID: 1,
			input: &dto.ApplicantProfileUpdate{
				City: "Москва",
			},
			mockSetup: func(applicantRepo *mock.MockApplicantRepository, cityRepo *mock.MockCityRepository) {
				cityRepo.EXPECT().
					GetCityByName(gomock.Any(), "Москва").
					Return(&entity.City{ID: 1}, nil)
				applicantRepo.EXPECT().
					UpdateApplicant(gomock.Any(), 1, map[string]interface{}{
						"city_id": 1,
					}).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Невалидное имя (слишком длинное)",
			userID: 1,
			input: &dto.ApplicantProfileUpdate{
				FirstName: strings.Repeat("a", 31),
			},
			mockSetup: func(*mock.MockApplicantRepository, *mock.MockCityRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных: first_name: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa does not validate as runelength(2|30)"),
			),
		},
		{
			name:   "Невалидная фамилия (слишком длинная)",
			userID: 1,
			input: &dto.ApplicantProfileUpdate{
				LastName: strings.Repeat("a", 31),
			},
			mockSetup: func(*mock.MockApplicantRepository, *mock.MockCityRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неправильный формат данных: last_name: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa does not validate as runelength(2|30)"),
			),
		},
		{
			name:   "Невалидная дата рождения",
			userID: 1,
			input: &dto.ApplicantProfileUpdate{
				BirthDate: time.Now().Add(24 * time.Hour),
			},
			mockSetup: func(*mock.MockApplicantRepository, *mock.MockCityRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("дата рождения не может быть позже текущей даты"),
			),
		},
		{
			name:   "Невалидный статус",
			userID: 1,
			input: &dto.ApplicantProfileUpdate{
				Status: "InvalidStatus",
			},
			mockSetup: func(*mock.MockApplicantRepository, *mock.MockCityRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("неверный статус соискателя"),
			),
		},
		{
			name:   "Город не найден",
			userID: 1,
			input: &dto.ApplicantProfileUpdate{
				City: "НесуществующийГород",
			},
			mockSetup: func(applicantRepo *mock.MockApplicantRepository, cityRepo *mock.MockCityRepository) {
				cityRepo.EXPECT().
					GetCityByName(gomock.Any(), "НесуществующийГород").
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("город с name=НесуществующийГород не найден"),
					),
					)
			},
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("город с name=НесуществующийГород не найден"),
			),
		},
		{
			name:   "Ошибка при обновлении",
			userID: 1,
			input: &dto.ApplicantProfileUpdate{
				FirstName: "Имя",
			},
			mockSetup: func(applicantRepo *mock.MockApplicantRepository, cityRepo *mock.MockCityRepository) {
				applicantRepo.EXPECT().
					UpdateApplicant(gomock.Any(), 1, map[string]interface{}{
						"first_name": "Имя",
					}).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось обновить соискателя с id=1"),
					),
					)
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось обновить соискателя с id=1"),
			),
		},
		{
			name:      "Нет полей для обновления",
			userID:    1,
			input:     &dto.ApplicantProfileUpdate{},
			mockSetup: func(*mock.MockApplicantRepository, *mock.MockCityRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("отсутствуют поля для обновления"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockApplicantRepo := mock.NewMockApplicantRepository(ctrl)
			mockCityRepo := mock.NewMockCityRepository(ctrl)
			applicantService := NewApplicantService(mockApplicantRepo, mockCityRepo, nil)

			tc.mockSetup(mockApplicantRepo, mockCityRepo)

			err := applicantService.UpdateProfile(context.Background(), tc.userID, tc.input)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var serviceErr entity.Error
				require.ErrorAs(t, err, &serviceErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestApplicantService_GetUser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		applicantID    int
		mockSetup      func(*mock.MockApplicantRepository, *mock.MockCityRepository)
		expectedResult *dto.ApplicantProfileResponse
		expectedErr    error
	}{
		{
			name:        "Успешное получение профиля",
			applicantID: 1,
			mockSetup: func(appRepo *mock.MockApplicantRepository, cityRepo *mock.MockCityRepository) {
				appRepo.EXPECT().
					GetApplicantByID(gomock.Any(), 1).
					Return(&entity.Applicant{
						ID:         1,
						FirstName:  "Иван",
						LastName:   "Иванов",
						MiddleName: "Иванович",
						BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
						Sex:        "M",
						Status:     "actively_searching",
						Quote:      "Тестовая цитата",
						Vk:         "https://vk.com/ivanov",
						Telegram:   "https://t.me/ivanov",
						Facebook:   "https://facebook.com/ivanov",
						CityID:     1,
					}, nil)

				cityRepo.EXPECT().
					GetCityByID(gomock.Any(), 1).
					Return(&entity.City{ID: 1, Name: "Москва"}, nil)
			},
			expectedResult: &dto.ApplicantProfileResponse{
				ID:         1,
				FirstName:  "Иван",
				LastName:   "Иванов",
				MiddleName: "Иванович",
				BirthDate:  time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
				Sex:        "M",
				Status:     "actively_searching",
				Quote:      "Тестовая цитата",
				Vk:         "https://vk.com/ivanov",
				Telegram:   "https://t.me/ivanov",
				Facebook:   "https://facebook.com/ivanov",
				City:       "Москва",
			},
			expectedErr: nil,
		},
		{
			name:        "Пользователь не найден",
			applicantID: 999,
			mockSetup: func(appRepo *mock.MockApplicantRepository, cityRepo *mock.MockCityRepository) {
				appRepo.EXPECT().
					GetApplicantByID(gomock.Any(), 999).
					Return(nil,
						entity.NewError(
							entity.ErrNotFound,
							fmt.Errorf("соискатель с id=999 не найден"),
						),
					)
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("соискатель с id=999 не найден"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAppRepo := mock.NewMockApplicantRepository(ctrl)
			mockCityRepo := mock.NewMockCityRepository(ctrl)
			applicantService := NewApplicantService(mockAppRepo, mockCityRepo, nil)

			tc.mockSetup(mockAppRepo, mockCityRepo)

			result, err := applicantService.GetUser(context.Background(), tc.applicantID)

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

func TestApplicantService_UpdateAvatar(t *testing.T) {
	t.Parallel()

	// Тестовые данные
	testImage := []byte{0x89, 0x50, 0x4E, 0x47} // Просто пример PNG заголовка
	mockAvatarResponse := &dto.UploadStaticResponse{ID: 100}

	testCases := []struct {
		name        string
		userID      int
		imageData   []byte
		mockSetup   func(*mock.MockApplicantRepository, *mockUC.MockStatic)
		expected    *dto.UploadStaticResponse
		expectedErr error
	}{
		{
			name:      "Успешное обновление аватара",
			userID:    1,
			imageData: testImage,
			mockSetup: func(appRepo *mock.MockApplicantRepository, staticUC *mockUC.MockStatic) {
				staticUC.EXPECT().
					UploadStatic(gomock.Any(), testImage).
					Return(mockAvatarResponse, nil)

				appRepo.EXPECT().
					GetApplicantByID(gomock.Any(), 1).
					Return(&entity.Applicant{AvatarID: 0}, nil)

				appRepo.EXPECT().
					UpdateApplicant(
						gomock.Any(),
						1,
						map[string]interface{}{"avatar_id": mockAvatarResponse.ID},
					).
					Return(nil)
			},
			expected:    mockAvatarResponse,
			expectedErr: nil,
		},
		{
			name:      "Ошибка загрузки аватара",
			userID:    2,
			imageData: testImage,
			mockSetup: func(appRepo *mock.MockApplicantRepository, staticUC *mockUC.MockStatic) {
				staticUC.EXPECT().
					UploadStatic(gomock.Any(), testImage).
					Return(nil, entity.NewError(entity.ErrInternal, fmt.Errorf("ошибка загрузки")))
			},
			expected:    nil,
			expectedErr: entity.NewError(entity.ErrInternal, fmt.Errorf("ошибка загрузки")),
		},
		{
			name:      "Ошибка при удалении старого аватара",
			userID:    3,
			imageData: testImage,
			mockSetup: func(appRepo *mock.MockApplicantRepository, staticUC *mockUC.MockStatic) {
				staticUC.EXPECT().
					UploadStatic(gomock.Any(), testImage).
					Return(mockAvatarResponse, nil)

				appRepo.EXPECT().
					GetApplicantByID(gomock.Any(), 3).
					Return(&entity.Applicant{AvatarID: 50}, nil)

				staticUC.EXPECT().
					DeleteStatic(gomock.Any(), 50).
					Return(entity.NewError(entity.ErrInternal, fmt.Errorf("ошибка удаления")))
			},
			expected:    nil,
			expectedErr: entity.NewError(entity.ErrInternal, fmt.Errorf("ошибка удаления")),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAppRepo := mock.NewMockApplicantRepository(ctrl)
			mockStaticUC := mockUC.NewMockStatic(ctrl)
			service := NewApplicantService(mockAppRepo, nil, mockStaticUC)

			tc.mockSetup(mockAppRepo, mockStaticUC)

			result, err := service.UpdateAvatar(context.Background(), tc.userID, tc.imageData)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestApplicantService_EmailExists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		email          string
		mockSetup      func(*mock.MockApplicantRepository)
		expectedResult *dto.EmailExistsResponse
		expectedErr    error
	}{
		{
			name:  "Успешный поиск email для соискателя",
			email: "applicant@example.com",
			mockSetup: func(appRepo *mock.MockApplicantRepository) {
				appRepo.EXPECT().
					GetApplicantByEmail(gomock.Any(), "applicant@example.com").
					Return(&entity.Applicant{ID: 1, Email: "applicant@example.com"}, nil)
			},
			expectedResult: &dto.EmailExistsResponse{
				Exists: true,
				Role:   "applicant",
			},
			expectedErr: nil,
		},
		{
			name:      "Неправильный формат почты",
			email:     "applicant_wrong_mail.com",
			mockSetup: func(m *mock.MockApplicantRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("невалидная почта"),
			),
		},
		{
			name:  "Email не найден",
			email: "nonexistent@example.com",
			mockSetup: func(appRepo *mock.MockApplicantRepository) {
				appRepo.EXPECT().
					GetApplicantByEmail(gomock.Any(), "nonexistent@example.com").
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("соискатель с email=nonexistent@example.com не найден"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("соискатель с email=nonexistent@example.com не найден"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAppRepo := mock.NewMockApplicantRepository(ctrl)
			service := NewApplicantService(mockAppRepo, nil, nil)

			tc.mockSetup(mockAppRepo)

			res, err := service.EmailExists(context.Background(), tc.email)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, res)
			}
		})
	}
}
