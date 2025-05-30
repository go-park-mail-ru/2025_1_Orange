package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository/mock"
	m "ResuMatch/internal/usecase/mock"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestChatService_StartChat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		vacancyID      int
		resumeID       int
		applicantID    int
		employerID     int
		mockSetup      func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository)
		expectedChatID int
		expectedErr    error
	}{
		{
			name:        "Success - chat created and message sent",
			vacancyID:   1,
			resumeID:    2,
			applicantID: 3,
			employerID:  4,
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository) {
				chatRepo.EXPECT().
					CreateChat(gomock.Any(), 1, 2, 4, 3).
					Return(&entity.Chat{ID: 10, ApplicantID: 3}, nil)

				messageRepo.EXPECT().
					CreateMessage(gomock.Any(), 10, 3, true, ResponseMessage).
					Return(&entity.Message{ID: 100}, nil)
			},
			expectedChatID: 10,
			expectedErr:    nil,
		},
		{
			name:        "Error - CreateChat fails",
			vacancyID:   5,
			resumeID:    6,
			applicantID: 7,
			employerID:  8,
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository) {
				chatRepo.EXPECT().
					CreateChat(gomock.Any(), 5, 6, 8, 7).
					Return(nil, errors.New("failed to create chat"))
				// CreateMessage НЕ вызывается
			},
			expectedChatID: -1,
			expectedErr:    errors.New("failed to create chat"),
		},
		{
			name:        "Error - CreateMessage fails",
			vacancyID:   9,
			resumeID:    10,
			applicantID: 11,
			employerID:  12,
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository) {
				chatRepo.EXPECT().
					CreateChat(gomock.Any(), 9, 10, 12, 11).
					Return(&entity.Chat{ID: 20, ApplicantID: 11}, nil)

				messageRepo.EXPECT().
					CreateMessage(gomock.Any(), 20, 11, true, ResponseMessage).
					Return(nil, errors.New("failed to create message"))
			},
			expectedChatID: -1,
			expectedErr:    errors.New("failed to create message"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockChatRepo := mock.NewMockChatRepository(ctrl)
			mockMessageRepo := mock.NewMockMessageRepository(ctrl)

			tc.mockSetup(mockChatRepo, mockMessageRepo)

			service := &ChatService{
				ChatRepo:    mockChatRepo,
				MessageRepo: mockMessageRepo,
			}

			ctx := context.Background()
			chatID, err := service.StartChat(ctx, tc.vacancyID, tc.resumeID, tc.applicantID, tc.employerID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Equal(t, -1, chatID)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedChatID, chatID)
			}
		})
	}
}

func TestChatService_GetChat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		chatID    int
		userID    int
		role      string
		mockSetup func(
			chatRepo *mock.MockChatRepository,
			vacancyUC *m.MockVacancy,
			resumeUC *m.MockResumeUsecase,
			applicantUC *m.MockApplicant,
			employerUC *m.MockEmployer,
		)
		expectedResult *dto.ChatResponse
		expectedErr    error
	}{
		{
			name:   "Error - GetChatByID fails",
			chatID: 2,
			userID: 20,
			role:   "employer",
			mockSetup: func(chatRepo *mock.MockChatRepository, vacancyUC *m.MockVacancy, resumeUC *m.MockResumeUsecase, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 2).
					Return(nil, errors.New("chat not found"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("chat not found"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockChatRepo := mock.NewMockChatRepository(ctrl)
			mockVacancyUC := m.NewMockVacancy(ctrl)
			mockResumeUC := m.NewMockResumeUsecase(ctrl)
			mockApplicantUC := m.NewMockApplicant(ctrl)
			mockEmployerUC := m.NewMockEmployer(ctrl)

			tc.mockSetup(mockChatRepo, mockVacancyUC, mockResumeUC, mockApplicantUC, mockEmployerUC)

			service := &ChatService{
				ChatRepo:    mockChatRepo,
				VacancyUC:   mockVacancyUC,
				ResumeUC:    mockResumeUC,
				ApplicantUC: mockApplicantUC,
				EmployerUC:  mockEmployerUC,
			}

			ctx := context.Background()
			resp, err := service.GetChat(ctx, tc.chatID, tc.userID, tc.role)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, resp)
			}
		})
	}
}

func TestChatService_SendMessage(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name      string
		chatID    int
		senderID  int
		role      string
		payload   string
		mockSetup func(
			chatRepo *mock.MockChatRepository,
			messageRepo *mock.MockMessageRepository,
			applicantUC *m.MockApplicant,
			employerUC *m.MockEmployer,
		)
		expectedResult *dto.MessageResponse
		expectedErr    error
	}{
		{
			name:     "Success - from applicant",
			chatID:   1,
			senderID: 10,
			role:     "applicant",
			payload:  "Hello from applicant",
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 1).
					Return(&entity.Chat{
						ID:          1,
						ApplicantID: 10,
						EmployerID:  20,
					}, nil)

				messageRepo.EXPECT().
					CreateMessage(gomock.Any(), 1, 10, true, "Hello from applicant").
					Return(&entity.Message{
						ID:            100,
						ChatID:        1,
						SenderID:      10,
						FromApplicant: true,
						Payload:       "Hello from applicant",
						SentAt:        now,
					}, nil)

				applicantUC.EXPECT().
					GetUser(gomock.Any(), 10).
					Return(&dto.ApplicantProfileResponse{
						AvatarPath: "/avatars/applicant10.png",
					}, nil)
			},
			expectedResult: &dto.MessageResponse{
				ID:            100,
				ChatID:        1,
				SenderID:      10,
				ReceiverID:    20,
				Avatar:        "/avatars/applicant10.png",
				FromApplicant: true,
				Payload:       "Hello from applicant",
				SentAt:        now,
			},
			expectedErr: nil,
		},
		{
			name:     "Success - from employer",
			chatID:   2,
			senderID: 20,
			role:     "employer",
			payload:  "Hello from employer",
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 2).
					Return(&entity.Chat{
						ID:          2,
						ApplicantID: 30,
						EmployerID:  20,
					}, nil)

				messageRepo.EXPECT().
					CreateMessage(gomock.Any(), 2, 20, false, "Hello from employer").
					Return(&entity.Message{
						ID:            101,
						ChatID:        2,
						SenderID:      20,
						FromApplicant: false,
						Payload:       "Hello from employer",
						SentAt:        now,
					}, nil)

				employerUC.EXPECT().
					GetUser(gomock.Any(), 20).
					Return(&dto.EmployerProfileResponse{
						LogoPath: "/logos/employer20.png",
					}, nil)
			},
			expectedResult: &dto.MessageResponse{
				ID:            101,
				ChatID:        2,
				SenderID:      20,
				ReceiverID:    30,
				Avatar:        "/logos/employer20.png",
				FromApplicant: false,
				Payload:       "Hello from employer",
				SentAt:        now,
			},
			expectedErr: nil,
		},
		{
			name:     "Error - GetChatByID fails",
			chatID:   3,
			senderID: 30,
			role:     "applicant",
			payload:  "Test",
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 3).
					Return(nil, errors.New("chat not found"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("chat not found"),
		},
		{
			name:     "Error - CreateMessage fails",
			chatID:   4,
			senderID: 40,
			role:     "employer",
			payload:  "Test message",
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 4).
					Return(&entity.Chat{
						ID:          4,
						ApplicantID: 50,
						EmployerID:  40,
					}, nil)

				messageRepo.EXPECT().
					CreateMessage(gomock.Any(), 4, 40, false, "Test message").
					Return(nil, errors.New("failed to create message"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("failed to create message"),
		},
		{
			name:     "Error - GetUser (applicant) fails",
			chatID:   5,
			senderID: 60,
			role:     "applicant",
			payload:  "Hi",
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 5).
					Return(&entity.Chat{
						ID:          5,
						ApplicantID: 60,
						EmployerID:  70,
					}, nil)

				messageRepo.EXPECT().
					CreateMessage(gomock.Any(), 5, 60, true, "Hi").
					Return(&entity.Message{
						ID:            102,
						ChatID:        5,
						SenderID:      60,
						FromApplicant: true,
						Payload:       "Hi",
						SentAt:        now,
					}, nil)

				applicantUC.EXPECT().
					GetUser(gomock.Any(), 60).
					Return(nil, errors.New("applicant not found"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("applicant not found"),
		},
		{
			name:     "Error - GetUser (employer) fails",
			chatID:   6,
			senderID: 70,
			role:     "employer",
			payload:  "Hello",
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 6).
					Return(&entity.Chat{
						ID:          6,
						ApplicantID: 80,
						EmployerID:  70,
					}, nil)

				messageRepo.EXPECT().
					CreateMessage(gomock.Any(), 6, 70, false, "Hello").
					Return(&entity.Message{
						ID:            103,
						ChatID:        6,
						SenderID:      70,
						FromApplicant: false,
						Payload:       "Hello",
						SentAt:        now,
					}, nil)

				employerUC.EXPECT().
					GetUser(gomock.Any(), 70).
					Return(nil, errors.New("employer not found"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("employer not found"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockChatRepo := mock.NewMockChatRepository(ctrl)
			mockMessageRepo := mock.NewMockMessageRepository(ctrl)
			mockApplicantUC := m.NewMockApplicant(ctrl)
			mockEmployerUC := m.NewMockEmployer(ctrl)

			tc.mockSetup(mockChatRepo, mockMessageRepo, mockApplicantUC, mockEmployerUC)

			service := &ChatService{
				ChatRepo:    mockChatRepo,
				MessageRepo: mockMessageRepo,
				ApplicantUC: mockApplicantUC,
				EmployerUC:  mockEmployerUC,
			}

			ctx := context.Background()
			resp, err := service.SendMessage(ctx, tc.chatID, tc.senderID, tc.role, tc.payload)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedErr.Error())
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, resp)
			}
		})
	}
}

func TestChatService_GetUserChats(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		userID      int
		role        string
		setupMocks  func(*mock.MockChatRepository, *m.MockVacancy, *m.MockEmployer, *m.MockApplicant)
		want        dto.ChatResponseList
		expectedErr error
	}

	tests := []testCase{
		{
			name:   "success for applicant",
			userID: 1,
			role:   "applicant",
			setupMocks: func(chatRepo *mock.MockChatRepository, vacancyUC *m.MockVacancy, employerUC *m.MockEmployer, applicantUC *m.MockApplicant) {
				chatRepo.EXPECT().GetForUser(gomock.Any(), 1, true).Return([]*entity.Chat{
					{ID: 10, VacancyID: 100, EmployerID: 200, ApplicantID: 1},
					{ID: 11, VacancyID: 101, EmployerID: 201, ApplicantID: 1},
				}, nil)

				vacancyUC.EXPECT().GetVacancy(gomock.Any(), 100, 1, "applicant").Return(&dto.VacancyResponse{ID: 100, EmployerID: 200, Title: "Vacancy 100"}, nil)
				vacancyUC.EXPECT().GetVacancy(gomock.Any(), 101, 1, "applicant").Return(&dto.VacancyResponse{ID: 101, EmployerID: 201, Title: "Vacancy 101"}, nil)

				employerUC.EXPECT().GetUser(gomock.Any(), 200).Return(&dto.EmployerProfileResponse{ID: 200, CompanyName: "Company A", LogoPath: "logoA.png"}, nil)
				employerUC.EXPECT().GetUser(gomock.Any(), 201).Return(&dto.EmployerProfileResponse{ID: 201, CompanyName: "Company B", LogoPath: "logoB.png"}, nil)
			},
			want: dto.ChatResponseList{
				{ID: 10, VacancyTitle: "Vacancy 100", User: dto.ChatUserPreview{ID: 200, Name: "Company A", AvatarPath: "logoA.png"}},
				{ID: 11, VacancyTitle: "Vacancy 101", User: dto.ChatUserPreview{ID: 201, Name: "Company B", AvatarPath: "logoB.png"}},
			},
		},
		{
			name:   "success for employer",
			userID: 2,
			role:   "employer",
			setupMocks: func(chatRepo *mock.MockChatRepository, vacancyUC *m.MockVacancy, employerUC *m.MockEmployer, applicantUC *m.MockApplicant) {
				chatRepo.EXPECT().GetForUser(gomock.Any(), 2, false).Return([]*entity.Chat{
					{ID: 20, VacancyID: 200, EmployerID: 2, ApplicantID: 300},
				}, nil)

				vacancyUC.EXPECT().GetVacancy(gomock.Any(), 200, 2, "employer").Return(&dto.VacancyResponse{ID: 200, EmployerID: 2, Title: "Vacancy 200"}, nil)
				applicantUC.EXPECT().GetUser(gomock.Any(), 300).Return(&dto.ApplicantProfileResponse{
					ID: 300, FirstName: "Ivan", LastName: "Ivanov", MiddleName: "Ivanovich", AvatarPath: "avatar.png",
				}, nil)
			},
			want: dto.ChatResponseList{
				{ID: 20, VacancyTitle: "Vacancy 200", User: dto.ChatUserPreview{ID: 300, Name: "Ivanov Ivan Ivanovich", AvatarPath: "avatar.png"}},
			},
		},
		{
			name:   "repo error",
			userID: 1,
			role:   "applicant",
			setupMocks: func(chatRepo *mock.MockChatRepository, _ *m.MockVacancy, _ *m.MockEmployer, _ *m.MockApplicant) {
				chatRepo.EXPECT().GetForUser(gomock.Any(), 1, true).Return(nil, errors.New("repo error"))
			},
			expectedErr: errors.New("repo error"),
		},
		{
			name:   "vacancy error",
			userID: 1,
			role:   "applicant",
			setupMocks: func(chatRepo *mock.MockChatRepository, vacancyUC *m.MockVacancy, _ *m.MockEmployer, _ *m.MockApplicant) {
				chatRepo.EXPECT().GetForUser(gomock.Any(), 1, true).Return([]*entity.Chat{
					{ID: 10, VacancyID: 100, EmployerID: 200, ApplicantID: 1},
				}, nil)

				vacancyUC.EXPECT().GetVacancy(gomock.Any(), 100, 1, "applicant").Return(nil, errors.New("vacancy error"))
			},
			expectedErr: errors.New("vacancy error"),
		},
		{
			name:   "employer get user error",
			userID: 1,
			role:   "applicant",
			setupMocks: func(chatRepo *mock.MockChatRepository, vacancyUC *m.MockVacancy, employerUC *m.MockEmployer, _ *m.MockApplicant) {
				chatRepo.EXPECT().GetForUser(gomock.Any(), 1, true).Return([]*entity.Chat{
					{ID: 10, VacancyID: 100, EmployerID: 200, ApplicantID: 1},
				}, nil)

				vacancyUC.EXPECT().GetVacancy(gomock.Any(), 100, 1, "applicant").Return(&dto.VacancyResponse{ID: 100, EmployerID: 200, Title: "Vacancy 100"}, nil)
				employerUC.EXPECT().GetUser(gomock.Any(), 200).Return(nil, errors.New("employer error"))
			},
			expectedErr: errors.New("employer error"),
		},
		{
			name:   "applicant get user error",
			userID: 2,
			role:   "employer",
			setupMocks: func(chatRepo *mock.MockChatRepository, vacancyUC *m.MockVacancy, _ *m.MockEmployer, applicantUC *m.MockApplicant) {
				chatRepo.EXPECT().GetForUser(gomock.Any(), 2, false).Return([]*entity.Chat{
					{ID: 20, VacancyID: 200, EmployerID: 2, ApplicantID: 300},
				}, nil)

				vacancyUC.EXPECT().GetVacancy(gomock.Any(), 200, 2, "employer").Return(&dto.VacancyResponse{ID: 200, EmployerID: 2, Title: "Vacancy 200"}, nil)
				applicantUC.EXPECT().GetUser(gomock.Any(), 300).Return(nil, errors.New("applicant error"))
			},
			expectedErr: errors.New("applicant error"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			chatRepo := mock.NewMockChatRepository(ctrl)
			vacancyUC := m.NewMockVacancy(ctrl)
			employerUC := m.NewMockEmployer(ctrl)
			applicantUC := m.NewMockApplicant(ctrl)

			service := &ChatService{
				ChatRepo:    chatRepo,
				VacancyUC:   vacancyUC,
				EmployerUC:  employerUC,
				ApplicantUC: applicantUC,
			}

			tc.setupMocks(chatRepo, vacancyUC, employerUC, applicantUC)

			got, err := service.GetUserChats(context.Background(), tc.userID, tc.role)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Nil(t, got)
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestChatService_GetChatMessages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := []struct {
		name      string
		chatID    int
		mockSetup func(
			chatRepo *mock.MockChatRepository,
			messageRepo *mock.MockMessageRepository,
			applicantUC *m.MockApplicant,
			employerUC *m.MockEmployer,
		)
		expectedResult dto.MessagesResponseList
		expectedErr    error
	}{
		{
			name:   "Success - messages from applicant and employer",
			chatID: 1,
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 1).
					Return(&entity.Chat{ID: 1, EmployerID: 10, ApplicantID: 20}, nil)

				messageRepo.EXPECT().
					GetMessagesForChat(gomock.Any(), 1).
					Return([]*entity.Message{
						{ID: 100, ChatID: 1, SenderID: 20, FromApplicant: true, Payload: "hello", SentAt: time.Now()},
						{ID: 101, ChatID: 1, SenderID: 10, FromApplicant: false, Payload: "hi", SentAt: time.Now()},
					}, nil)

				applicantUC.EXPECT().
					GetUser(gomock.Any(), 20).
					Return(&dto.ApplicantProfileResponse{ID: 20, AvatarPath: "/avatars/applicant.png"}, nil)

				employerUC.EXPECT().
					GetUser(gomock.Any(), 10).
					Return(&dto.EmployerProfileResponse{ID: 10, LogoPath: "/logos/employer.png"}, nil)
			},
			expectedResult: dto.MessagesResponseList{
				{
					ID:            100,
					ChatID:        1,
					SenderID:      20,
					ReceiverID:    10,
					Avatar:        "/avatars/applicant.png",
					FromApplicant: true,
					Payload:       "hello",
				},
				{
					ID:            101,
					ChatID:        1,
					SenderID:      10,
					ReceiverID:    20,
					Avatar:        "/logos/employer.png",
					FromApplicant: false,
					Payload:       "hi",
				},
			},
			expectedErr: nil,
		},
		{
			name:   "Error - chat not found",
			chatID: 2,
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 2).
					Return(nil, errors.New("chat not found"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("chat not found"),
		},
		{
			name:   "Error - failed to get messages",
			chatID: 3,
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 3).
					Return(&entity.Chat{ID: 3, EmployerID: 10, ApplicantID: 20}, nil)

				messageRepo.EXPECT().
					GetMessagesForChat(gomock.Any(), 3).
					Return(nil, errors.New("db error"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("db error"),
		},
		{
			name:   "Error - failed to get applicant user",
			chatID: 4,
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 4).
					Return(&entity.Chat{ID: 4, EmployerID: 10, ApplicantID: 20}, nil)

				messageRepo.EXPECT().
					GetMessagesForChat(gomock.Any(), 4).
					Return([]*entity.Message{
						{ID: 200, ChatID: 4, SenderID: 20, FromApplicant: true, Payload: "hey", SentAt: time.Now()},
					}, nil)

				applicantUC.EXPECT().
					GetUser(gomock.Any(), 20).
					Return(nil, errors.New("applicant not found"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("applicant not found"),
		},
		{
			name:   "Error - failed to get employer user",
			chatID: 5,
			mockSetup: func(chatRepo *mock.MockChatRepository, messageRepo *mock.MockMessageRepository, applicantUC *m.MockApplicant, employerUC *m.MockEmployer) {
				chatRepo.EXPECT().
					GetChatByID(gomock.Any(), 5).
					Return(&entity.Chat{ID: 5, EmployerID: 10, ApplicantID: 20}, nil)

				messageRepo.EXPECT().
					GetMessagesForChat(gomock.Any(), 5).
					Return([]*entity.Message{
						{ID: 201, ChatID: 5, SenderID: 10, FromApplicant: false, Payload: "hey", SentAt: time.Now()},
					}, nil)

				employerUC.EXPECT().
					GetUser(gomock.Any(), 10).
					Return(nil, errors.New("employer not found"))
			},
			expectedResult: nil,
			expectedErr:    errors.New("employer not found"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockChatRepo := mock.NewMockChatRepository(ctrl)
			mockMessageRepo := mock.NewMockMessageRepository(ctrl)
			mockApplicantUC := m.NewMockApplicant(ctrl)
			mockEmployerUC := m.NewMockEmployer(ctrl)

			tc.mockSetup(mockChatRepo, mockMessageRepo, mockApplicantUC, mockEmployerUC)

			service := &ChatService{
				ChatRepo:    mockChatRepo,
				MessageRepo: mockMessageRepo,
				ApplicantUC: mockApplicantUC,
				EmployerUC:  mockEmployerUC,
			}

			resp, err := service.GetChatMessages(ctx, tc.chatID)

			if tc.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErr.Error())
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResult), len(resp))
				for i := range resp {
					require.Equal(t, tc.expectedResult[i].ID, resp[i].ID)
					require.Equal(t, tc.expectedResult[i].ChatID, resp[i].ChatID)
					require.Equal(t, tc.expectedResult[i].SenderID, resp[i].SenderID)
					require.Equal(t, tc.expectedResult[i].ReceiverID, resp[i].ReceiverID)
					require.Equal(t, tc.expectedResult[i].Avatar, resp[i].Avatar)
					require.Equal(t, tc.expectedResult[i].FromApplicant, resp[i].FromApplicant)
					require.Equal(t, tc.expectedResult[i].Payload, resp[i].Payload)
				}
			}
		})
	}
}
