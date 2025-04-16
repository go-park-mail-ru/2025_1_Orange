// Code generated by MockGen. DO NOT EDIT.
// Source: ResuMatch/internal/repository (interfaces: SpecializationRepository)
//
// Generated by this command:
//
//	mockgen -package mock -destination internal/repository/mock/mock_specialization.go ResuMatch/internal/repository SpecializationRepository
//

// Package mock is a generated GoMock package.
package mock

import (
	entity "ResuMatch/internal/entity"
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockSpecializationRepository is a mock of SpecializationRepository interface.
type MockSpecializationRepository struct {
	ctrl     *gomock.Controller
	recorder *MockSpecializationRepositoryMockRecorder
	isgomock struct{}
}

// MockSpecializationRepositoryMockRecorder is the mock recorder for MockSpecializationRepository.
type MockSpecializationRepositoryMockRecorder struct {
	mock *MockSpecializationRepository
}

// NewMockSpecializationRepository creates a new mock instance.
func NewMockSpecializationRepository(ctrl *gomock.Controller) *MockSpecializationRepository {
	mock := &MockSpecializationRepository{ctrl: ctrl}
	mock.recorder = &MockSpecializationRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSpecializationRepository) EXPECT() *MockSpecializationRepositoryMockRecorder {
	return m.recorder
}

// GetAll mocks base method.
func (m *MockSpecializationRepository) GetAll(ctx context.Context) ([]entity.Specialization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx)
	ret0, _ := ret[0].([]entity.Specialization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockSpecializationRepositoryMockRecorder) GetAll(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockSpecializationRepository)(nil).GetAll), ctx)
}

// GetByID mocks base method.
func (m *MockSpecializationRepository) GetByID(ctx context.Context, id int) (*entity.Specialization, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*entity.Specialization)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockSpecializationRepositoryMockRecorder) GetByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockSpecializationRepository)(nil).GetByID), ctx, id)
}
