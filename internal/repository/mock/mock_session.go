// Code generated by MockGen. DO NOT EDIT.
// Source: ResuMatch/internal/repository (interfaces: SessionRepository)
//
// Generated by this command:
//
//	mockgen -package mock -destination internal/repository/mock/mock_session.go ResuMatch/internal/repository SessionRepository
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockSessionRepository is a mock of SessionRepository interface.
type MockSessionRepository struct {
	ctrl     *gomock.Controller
	recorder *MockSessionRepositoryMockRecorder
	isgomock struct{}
}

// MockSessionRepositoryMockRecorder is the mock recorder for MockSessionRepository.
type MockSessionRepositoryMockRecorder struct {
	mock *MockSessionRepository
}

// NewMockSessionRepository creates a new mock instance.
func NewMockSessionRepository(ctrl *gomock.Controller) *MockSessionRepository {
	mock := &MockSessionRepository{ctrl: ctrl}
	mock.recorder = &MockSessionRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSessionRepository) EXPECT() *MockSessionRepositoryMockRecorder {
	return m.recorder
}

// CreateSession mocks base method.
func (m *MockSessionRepository) CreateSession(ctx context.Context, userID int, role string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSession", ctx, userID, role)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSession indicates an expected call of CreateSession.
func (mr *MockSessionRepositoryMockRecorder) CreateSession(ctx, userID, role any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSession", reflect.TypeOf((*MockSessionRepository)(nil).CreateSession), ctx, userID, role)
}

// DeleteAllSessions mocks base method.
func (m *MockSessionRepository) DeleteAllSessions(ctx context.Context, userID int, role string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAllSessions", ctx, userID, role)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAllSessions indicates an expected call of DeleteAllSessions.
func (mr *MockSessionRepositoryMockRecorder) DeleteAllSessions(ctx, userID, role any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAllSessions", reflect.TypeOf((*MockSessionRepository)(nil).DeleteAllSessions), ctx, userID, role)
}

// DeleteSession mocks base method.
func (m *MockSessionRepository) DeleteSession(ctx context.Context, sessionToken string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSession", ctx, sessionToken)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSession indicates an expected call of DeleteSession.
func (mr *MockSessionRepositoryMockRecorder) DeleteSession(ctx, sessionToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSession", reflect.TypeOf((*MockSessionRepository)(nil).DeleteSession), ctx, sessionToken)
}

// GetSession mocks base method.
func (m *MockSessionRepository) GetSession(ctx context.Context, sessionToken string) (int, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSession", ctx, sessionToken)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetSession indicates an expected call of GetSession.
func (mr *MockSessionRepositoryMockRecorder) GetSession(ctx, sessionToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSession", reflect.TypeOf((*MockSessionRepository)(nil).GetSession), ctx, sessionToken)
}
