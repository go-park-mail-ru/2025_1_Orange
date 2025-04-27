package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/repository/mock"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestAuthService_CreateSession(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		userID      int
		role        string
		expected    string
		mockSetup   func(*mock.MockSessionRepository)
		expectedErr error
	}{
		{
			name:     "Успешное создание сессии",
			userID:   1,
			role:     "applicant",
			expected: "session_token_123",
			mockSetup: func(mockRepo *mock.MockSessionRepository) {
				mockRepo.EXPECT().
					CreateSession(gomock.Any(), 1, "applicant").
					Return("session_token_123", nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка при создании сессии",
			userID: 2,
			role:   "employer",
			mockSetup: func(mockRepo *mock.MockSessionRepository) {
				mockRepo.EXPECT().
					CreateSession(gomock.Any(), 2, "employer").
					Return("", entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось создать сессию для пользователя с id=2, role=employer"),
					))
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось создать сессию для пользователя с id=2, role=employer"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessRepo := mock.NewMockSessionRepository(ctrl)
			service := NewAuthService(mockSessRepo)

			tc.mockSetup(mockSessRepo)

			result, err := service.CreateSession(context.Background(), tc.userID, tc.role)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var serviceErr entity.Error
				require.ErrorAs(t, err, &serviceErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestAuthService_GetUserIDBySession(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		session      string
		mockSetup    func(*mock.MockSessionRepository)
		expectedID   int
		expectedRole string
		expectedErr  error
	}{
		{
			name:    "Сессия найдена",
			session: "valid_token",
			mockSetup: func(repo *mock.MockSessionRepository) {
				repo.EXPECT().
					GetSession(gomock.Any(), "valid_token").
					Return(1, "applicant", nil)
			},
			expectedID:   1,
			expectedRole: "applicant",
			expectedErr:  nil,
		},
		{
			name:    "Ошибка при получении сессии",
			session: "invalid_token",
			mockSetup: func(repo *mock.MockSessionRepository) {
				repo.EXPECT().
					GetSession(gomock.Any(), "invalid_token").
					Return(-1, "", entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось получить сессию с токеном=invalid_token"),
					))
			},
			expectedID:   -1,
			expectedRole: "",
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось получить сессию с токеном=invalid_token"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessRepo := mock.NewMockSessionRepository(ctrl)
			service := NewAuthService(mockSessRepo)

			tc.mockSetup(mockSessRepo)

			id, role, err := service.GetUserIDBySession(context.Background(), tc.session)

			if tc.expectedErr != nil {
				require.Error(t, err)
				var serviceErr entity.Error
				require.ErrorAs(t, err, &serviceErr)
				require.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedID, id)
				require.Equal(t, tc.expectedRole, role)
			}
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		session     string
		mockSetup   func(*mock.MockSessionRepository)
		expectedErr error
	}{
		{
			name:    "Успешный выход",
			session: "logout_token",
			mockSetup: func(repo *mock.MockSessionRepository) {
				repo.EXPECT().
					DeleteSession(gomock.Any(), "logout_token").
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:    "Ошибка при выходе",
			session: "bad_token",
			mockSetup: func(repo *mock.MockSessionRepository) {
				repo.EXPECT().
					DeleteSession(gomock.Any(), "bad_token").
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось удалить сессию с токеном=bad_token"),
					))
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось удалить сессию с токеном=bad_token"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessRepo := mock.NewMockSessionRepository(ctrl)
			service := NewAuthService(mockSessRepo)

			tc.mockSetup(mockSessRepo)

			err := service.Logout(context.Background(), tc.session)

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

func TestAuthService_LogoutAll(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		userID      int
		role        string
		mockSetup   func(*mock.MockSessionRepository)
		expectedErr error
	}{
		{
			name:   "Успешный выход со всех устройств",
			userID: 10,
			role:   "employer",
			mockSetup: func(repo *mock.MockSessionRepository) {
				repo.EXPECT().
					DeleteAllSessions(gomock.Any(), 10, "employer").
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:   "Ошибка при выходе со всех устройств",
			userID: 99,
			role:   "applicant",
			mockSetup: func(repo *mock.MockSessionRepository) {
				repo.EXPECT().
					DeleteAllSessions(gomock.Any(), 99, "applicant").
					Return(entity.NewError(
						entity.ErrInternal,
						fmt.Errorf("не удалось удалить активные сессии пользователя"),
					))
			},
			expectedErr: entity.NewError(
				entity.ErrInternal,
				fmt.Errorf("не удалось удалить активные сессии пользователя"),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSessRepo := mock.NewMockSessionRepository(ctrl)
			service := NewAuthService(mockSessRepo)

			tc.mockSetup(mockSessRepo)

			err := service.LogoutAll(context.Background(), tc.userID, tc.role)

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
