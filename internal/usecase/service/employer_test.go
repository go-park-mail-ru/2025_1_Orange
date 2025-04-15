package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"strings"
	"testing"
	"time"
)

func TestEmployerService_Register(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       *dto.EmployerRegister
		mockSetup   func(*mock.MockEmployerRepository)
		expectedID  int
		expectedErr error
	}{
		{
			name: "Успешная регистрация",
			input: &dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "employer@email.com",
					Password: "StrongPassword1!",
				},
				CompanyName:  "Вконтакте",
				LegalAddress: "Москва, ул. Яблочкова",
			},
			mockSetup: func(m *mock.MockEmployerRepository) {
				m.EXPECT().
					CreateEmployer(
						gomock.Any(),
						"employer@email.com",
						"Вконтакте",
						"Москва, ул. Яблочкова",
						gomock.Any(), // hash
						gomock.Any(), // salt
					).
					Return(&entity.Employer{ID: 10}, nil)
			},
			expectedID:  10,
			expectedErr: nil,
		},
		{
			name: "Невалидная почта",
			input: &dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "invalid-email",
					Password: "StrongPassword1!",
				},
				CompanyName:  "Company",
				LegalAddress: "Address",
			},
			mockSetup:  func(m *mock.MockEmployerRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("невалидная почта"),
			),
		},
		{
			name: "Слабый пароль",
			input: &dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "employer@email.com",
					Password: "123",
				},
				CompanyName:  "Company",
				LegalAddress: "Address",
			},
			mockSetup:  func(m *mock.MockEmployerRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("пароль должен содержать не менее 8 символов"),
			),
		},
		{
			name: "Слишком длинное название компании",
			input: &dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "employer@email.com",
					Password: "StrongPassword1!",
				},
				CompanyName:  strings.Repeat("A", 65),
				LegalAddress: "Address",
			},
			mockSetup:  func(m *mock.MockEmployerRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("название компании не может быть длиннее 64 символов"),
			),
		},
		{
			name: "Слишком длинный юридический адрес",
			input: &dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "employer@email.com",
					Password: "StrongPassword1!",
				},
				CompanyName:  "Company",
				LegalAddress: strings.Repeat("B", 256),
			},
			mockSetup:  func(m *mock.MockEmployerRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("юридический адрес компании не может быть длиннее 255 символов"),
			),
		},
		{
			name: "Ошибка при создании работодателя",
			input: &dto.EmployerRegister{
				AuthCredentials: dto.AuthCredentials{
					Email:    "employer@email.com",
					Password: "StrongPassword1!",
				},
				CompanyName:  "Company",
				LegalAddress: "Address",
			},
			mockSetup: func(m *mock.MockEmployerRepository) {
				m.EXPECT().
					CreateEmployer(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					Return(nil, entity.NewError(
						entity.ErrInternal,
						errors.New("неизвестная ошибка при создании работодателя err=pq: test pq error"),
					))
			},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				errors.New("неизвестная ошибка при создании работодателя err=pq: test pq error"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockEmployerRepository(ctrl)
			employerService := NewEmployerService(mockRepo, nil)

			tc.mockSetup(mockRepo)

			id, err := employerService.Register(context.Background(), tc.input)

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

func TestEmployerService_Login(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       *dto.Login
		mockSetup   func(*mock.MockEmployerRepository)
		expectedID  int
		expectedErr error
	}{
		{
			name: "Успешная авторизация",
			input: &dto.Login{
				Email:    "valid@email.com",
				Password: "validpasswordI!",
			},
			mockSetup: func(m *mock.MockEmployerRepository) {
				salt, hash, _ := entity.HashPassword("validpasswordI!")
				m.EXPECT().
					GetEmployerByEmail(gomock.Any(), "valid@email.com").
					Return(&entity.Employer{
						ID:           1,
						PasswordHash: hash,
						PasswordSalt: salt,
					}, nil)
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name: "Невалидная почта",
			input: &dto.Login{
				Email:    "invalid_email.com",
				Password: "validpasswordI!",
			},
			mockSetup:  func(m *mock.MockEmployerRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("невалидная почта"),
			),
		},
		{
			name: "Невалидный пароль",
			input: &dto.Login{
				Email:    "valid@email.com",
				Password: "short",
			},
			mockSetup:  func(m *mock.MockEmployerRepository) {},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("пароль должен содержать не менее 8 символов"),
			),
		},
		{
			name: "Работодатель не найден",
			input: &dto.Login{
				Email:    "notfound@email.com",
				Password: "anypassword",
			},
			mockSetup: func(m *mock.MockEmployerRepository) {
				m.EXPECT().
					GetEmployerByEmail(gomock.Any(), "notfound@email.com").
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("работодатель с email=notfound@email.com не найден"),
					),
					)
			},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("работодатель с email=notfound@email.com не найден"),
			),
		},
		{
			name: "Неверный пароль",
			input: &dto.Login{
				Email:    "valid@email.com",
				Password: "wrongPassword123!",
			},
			mockSetup: func(m *mock.MockEmployerRepository) {
				salt, hash, _ := entity.HashPassword("correctPassword123!")
				m.EXPECT().
					GetEmployerByEmail(gomock.Any(), "valid@email.com").
					Return(&entity.Employer{
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
			name: "Ошибка получения работодателя",
			input: &dto.Login{
				Email:    "valid@email.com",
				Password: "validpasswordI!",
			},
			mockSetup: func(m *mock.MockEmployerRepository) {
				m.EXPECT().
					GetEmployerByEmail(gomock.Any(), "valid@email.com").
					Return(nil, entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось найти работодателя с email=valid@email.com"),
					),
					)
			},
			expectedID: -1,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось найти работодателя с email=valid@email.com"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock.NewMockEmployerRepository(ctrl)
			employerService := NewEmployerService(mockRepo, nil)

			tc.mockSetup(mockRepo)

			id, err := employerService.Login(context.Background(), tc.input)

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

func TestEmployerService_GetUser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		employerID     int
		mockSetup      func(*mock.MockEmployerRepository, *mock.MockStaticRepository)
		expectedResult *dto.EmployerProfileResponse
		expectedErr    error
	}{
		{
			name:       "Успешное получение профиля",
			employerID: 1,
			mockSetup: func(empRepo *mock.MockEmployerRepository, staticRepo *mock.MockStaticRepository) {
				empRepo.EXPECT().
					GetEmployerByID(gomock.Any(), 1).
					Return(&entity.Employer{
						ID:           1,
						CompanyName:  "Вконтакте",
						LegalAddress: "г. Москва, ул. Ленина, д. 1",
						Email:        "hr@vk.ru",
						Slogan:       "Лучшие кадры — к нам!",
						Website:      "https://vk.ru",
						Description:  "Разрабатываем множество крутых продуктов",
						Vk:           "https://vk.com/vk",
						Telegram:     "https://t.me/vk",
						Facebook:     "https://facebook.com/vk",
						LogoID:       5,
						CreatedAt:    time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					}, nil)

				staticRepo.EXPECT().
					GetStatic(gomock.Any(), 5).
					Return("assets/logo.png", nil)
			},
			expectedResult: &dto.EmployerProfileResponse{
				ID:           1,
				CompanyName:  "Вконтакте",
				LegalAddress: "г. Москва, ул. Ленина, д. 1",
				Email:        "hr@vk.ru",
				Slogan:       "Лучшие кадры — к нам!",
				Website:      "https://vk.ru",
				Description:  "Разрабатываем множество крутых продуктов",
				Vk:           "https://vk.com/vk",
				Telegram:     "https://t.me/vk",
				Facebook:     "https://facebook.com/vk",
				LogoPath:     "assets/logo.png",
				CreatedAt:    time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expectedErr: nil,
		},
		{
			name:       "Работодатель не найден",
			employerID: 999,
			mockSetup: func(empRepo *mock.MockEmployerRepository, staticRepo *mock.MockStaticRepository) {
				empRepo.EXPECT().
					GetEmployerByID(gomock.Any(), 999).
					Return(nil, entity.NewError(
						entity.ErrNotFound,
						fmt.Errorf("работодатель с id=999 не найден"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrNotFound,
				fmt.Errorf("работодатель с id=999 не найден"),
			),
		},
		{
			name:       "Ошибка получения логотипа",
			employerID: 2,
			mockSetup: func(empRepo *mock.MockEmployerRepository, staticRepo *mock.MockStaticRepository) {
				empRepo.EXPECT().
					GetEmployerByID(gomock.Any(), 2).
					Return(&entity.Employer{
						ID:           2,
						CompanyName:  "Компания",
						LegalAddress: "Москва",
						Email:        "email@company.com",
						LogoID:       10,
					}, nil)

				staticRepo.EXPECT().
					GetStatic(gomock.Any(), 10).
					Return("", entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось загрузить логотип"),
					))
			},
			expectedResult: nil,
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось загрузить логотип"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEmpRepo := mock.NewMockEmployerRepository(ctrl)
			mockStaticRepo := mock.NewMockStaticRepository(ctrl)
			employerService := NewEmployerService(mockEmpRepo, mockStaticRepo)

			tc.mockSetup(mockEmpRepo, mockStaticRepo)

			result, err := employerService.GetUser(context.Background(), tc.employerID)

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

func TestEmployerService_UpdateProfile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		userID      int
		input       *dto.EmployerProfileUpdate
		mockSetup   func(*mock.MockEmployerRepository)
		expectedErr error
	}{
		{
			name:   "Успешное обновление всех полей",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				CompanyName:  "Яндекс",
				LegalAddress: "Москва, ул. Льва Толстого, д. 16",
				Slogan:       "Найдётся всё",
				Website:      "https://yandex.ru",
				Description:  "Крупнейшая российская IT-компания, специализирующаяся на поиске, навигации и других интернет-сервисах.",
				Vk:           "https://vk.com/yandex",
				Telegram:     "https://t.me/yandex",
				Facebook:     "https://facebook.com/yandex",
			},
			mockSetup: func(employerRepo *mock.MockEmployerRepository) {
				employerRepo.EXPECT().
					UpdateEmployer(gomock.Any(), 1, map[string]interface{}{
						"company_name":  "Яндекс",
						"legal_address": "Москва, ул. Льва Толстого, д. 16",
						"slogan":        "Найдётся всё",
						"website":       "https://yandex.ru",
						"description":   "Крупнейшая российская IT-компания, специализирующаяся на поиске, навигации и других интернет-сервисах.",
						"vk":            "https://vk.com/yandex",
						"telegram":      "https://t.me/yandex",
						"facebook":      "https://facebook.com/yandex",
					}).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Невалидное название компании",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				CompanyName: strings.Repeat("a", 65),
			},
			mockSetup:   func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(entity.ErrBadRequest, fmt.Errorf("название компании не может быть длиннее 64 символов")),
		},
		{
			name:   "Невалидный юридический адрес",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				LegalAddress: strings.Repeat("b", 256),
			},
			mockSetup:   func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(entity.ErrBadRequest, fmt.Errorf("юридический адрес компании не может быть длиннее 255 символов")),
		},
		{
			name:   "Невалидный слоган",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				Slogan: strings.Repeat("c", 256),
			},
			mockSetup:   func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(entity.ErrBadRequest, fmt.Errorf("слоган компании не может быть длиннее 255 символов")),
		},
		{
			name:   "Невалидный сайт",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				Website: strings.Repeat("d", 129),
			},
			mockSetup:   func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(entity.ErrBadRequest, fmt.Errorf("url не может быть длиннее 128 символов")),
		},
		{
			name:   "Невалидная ссылка VK (слишком длинная)",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				Vk: strings.Repeat("a", 129),
			},
			mockSetup: func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("url не может быть длиннее 128 символов"),
			),
		},
		{
			name:   "Невалидная ссылка Telegram (слишком длинная)",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				Telegram: strings.Repeat("a", 129),
			},
			mockSetup: func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("url не может быть длиннее 128 символов"),
			),
		},
		{
			name:   "Невалидная ссылка Facebook (слишком длинная)",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				Facebook: strings.Repeat("a", 129),
			},
			mockSetup: func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("url не может быть длиннее 128 символов"),
			),
		},
		{
			name:   "Невалидное описание",
			userID: 1,
			input: &dto.EmployerProfileUpdate{
				Description: strings.Repeat("e", 2001),
			},
			mockSetup:   func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(entity.ErrBadRequest, fmt.Errorf("описание компании не может быть длиннее 2000 символов")),
		},
		{
			name:      "Нет полей для обновления",
			userID:    1,
			input:     &dto.EmployerProfileUpdate{},
			mockSetup: func(*mock.MockEmployerRepository) {},
			expectedErr: entity.NewError(
				entity.ErrBadRequest,
				fmt.Errorf("no fields to update"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			employerRepo := mock.NewMockEmployerRepository(ctrl)

			tc.mockSetup(employerRepo)

			employerService := NewEmployerService(employerRepo, nil)

			err := employerService.UpdateProfile(context.Background(), tc.userID, tc.input)

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

func TestEmployerService_UpdateLogo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		userID      int
		logoID      int
		mockSetup   func(*mock.MockEmployerRepository)
		expectedErr error
	}{
		{
			name:   "Успешное обновление логотипа",
			userID: 1,
			logoID: 100,
			mockSetup: func(empRepo *mock.MockEmployerRepository) {
				empRepo.EXPECT().
					UpdateEmployer(
						gomock.Any(),
						1,
						map[string]interface{}{"logo_id": 100},
					).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка при обновлении логотипа",
			userID: 2,
			logoID: 200,
			mockSetup: func(empRepo *mock.MockEmployerRepository) {
				empRepo.EXPECT().
					UpdateEmployer(
						gomock.Any(),
						2,
						map[string]interface{}{"logo_id": 200},
					).
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось обновить работодателя с id=2"),
					))
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось обновить работодателя с id=2"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockEmployerRepo := mock.NewMockEmployerRepository(ctrl)
			service := NewEmployerService(mockEmployerRepo, nil)

			tc.mockSetup(mockEmployerRepo)

			err := service.UpdateLogo(context.Background(), tc.userID, tc.logoID)

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
