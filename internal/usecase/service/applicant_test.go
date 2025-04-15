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
				fmt.Errorf("имя не может быть длиннее 30 символов"),
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
				fmt.Errorf("фамилия не может быть длиннее 30 символов"),
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
