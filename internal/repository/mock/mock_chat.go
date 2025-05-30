// Code generated by MockGen. DO NOT EDIT.
// Source: ResuMatch/internal/repository (interfaces: ChatRepository)
//
// Generated by this command:
//
//	mockgen -package mock -destination internal/repository/mock/mock_chat.go ResuMatch/internal/repository ChatRepository
//

// Package mock is a generated GoMock package.
package mock

import (
	entity "ResuMatch/internal/entity"
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockChatRepository is a mock of ChatRepository interface.
type MockChatRepository struct {
	ctrl     *gomock.Controller
	recorder *MockChatRepositoryMockRecorder
	isgomock struct{}
}

// MockChatRepositoryMockRecorder is the mock recorder for MockChatRepository.
type MockChatRepositoryMockRecorder struct {
	mock *MockChatRepository
}

// NewMockChatRepository creates a new mock instance.
func NewMockChatRepository(ctrl *gomock.Controller) *MockChatRepository {
	mock := &MockChatRepository{ctrl: ctrl}
	mock.recorder = &MockChatRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChatRepository) EXPECT() *MockChatRepositoryMockRecorder {
	return m.recorder
}

// CreateChat mocks base method.
func (m *MockChatRepository) CreateChat(ctx context.Context, vacancyID, resumeID, employerID, applicantID int) (*entity.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateChat", ctx, vacancyID, resumeID, employerID, applicantID)
	ret0, _ := ret[0].(*entity.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateChat indicates an expected call of CreateChat.
func (mr *MockChatRepositoryMockRecorder) CreateChat(ctx, vacancyID, resumeID, employerID, applicantID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateChat", reflect.TypeOf((*MockChatRepository)(nil).CreateChat), ctx, vacancyID, resumeID, employerID, applicantID)
}

// GetChatByID mocks base method.
func (m *MockChatRepository) GetChatByID(ctx context.Context, chatID int) (*entity.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChatByID", ctx, chatID)
	ret0, _ := ret[0].(*entity.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetChatByID indicates an expected call of GetChatByID.
func (mr *MockChatRepositoryMockRecorder) GetChatByID(ctx, chatID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChatByID", reflect.TypeOf((*MockChatRepository)(nil).GetChatByID), ctx, chatID)
}

// GetForUser mocks base method.
func (m *MockChatRepository) GetForUser(ctx context.Context, userID int, isApplicant bool) ([]*entity.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetForUser", ctx, userID, isApplicant)
	ret0, _ := ret[0].([]*entity.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetForUser indicates an expected call of GetForUser.
func (mr *MockChatRepositoryMockRecorder) GetForUser(ctx, userID, isApplicant any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetForUser", reflect.TypeOf((*MockChatRepository)(nil).GetForUser), ctx, userID, isApplicant)
}

// GetForVacancy mocks base method.
func (m *MockChatRepository) GetForVacancy(ctx context.Context, vacancyID, applicantID int) (*entity.Chat, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetForVacancy", ctx, vacancyID, applicantID)
	ret0, _ := ret[0].(*entity.Chat)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetForVacancy indicates an expected call of GetForVacancy.
func (mr *MockChatRepositoryMockRecorder) GetForVacancy(ctx, vacancyID, applicantID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetForVacancy", reflect.TypeOf((*MockChatRepository)(nil).GetForVacancy), ctx, vacancyID, applicantID)
}

// GetVacancyChatInfo mocks base method.
func (m *MockChatRepository) GetVacancyChatInfo(ctx context.Context, vacancyID, applicantID int) (*entity.VacancyChatInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVacancyChatInfo", ctx, vacancyID, applicantID)
	ret0, _ := ret[0].(*entity.VacancyChatInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetVacancyChatInfo indicates an expected call of GetVacancyChatInfo.
func (mr *MockChatRepositoryMockRecorder) GetVacancyChatInfo(ctx, vacancyID, applicantID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVacancyChatInfo", reflect.TypeOf((*MockChatRepository)(nil).GetVacancyChatInfo), ctx, vacancyID, applicantID)
}
