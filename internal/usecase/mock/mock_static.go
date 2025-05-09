// Code generated by MockGen. DO NOT EDIT.
// Source: ResuMatch/internal/usecase (interfaces: Static)
//
// Generated by this command:
//
//	mockgen -package mock -destination internal/usecase/mock/mock_static.go ResuMatch/internal/usecase Static
//

// Package mock is a generated GoMock package.
package mock

import (
	dto "ResuMatch/internal/entity/dto"
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockStatic is a mock of Static interface.
type MockStatic struct {
	ctrl     *gomock.Controller
	recorder *MockStaticMockRecorder
	isgomock struct{}
}

// MockStaticMockRecorder is the mock recorder for MockStatic.
type MockStaticMockRecorder struct {
	mock *MockStatic
}

// NewMockStatic creates a new mock instance.
func NewMockStatic(ctrl *gomock.Controller) *MockStatic {
	mock := &MockStatic{ctrl: ctrl}
	mock.recorder = &MockStaticMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStatic) EXPECT() *MockStaticMockRecorder {
	return m.recorder
}

// DeleteStatic mocks base method.
func (m *MockStatic) DeleteStatic(ctx context.Context, id int) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteStatic", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteStatic indicates an expected call of DeleteStatic.
func (mr *MockStaticMockRecorder) DeleteStatic(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteStatic", reflect.TypeOf((*MockStatic)(nil).DeleteStatic), ctx, id)
}

// GetStatic mocks base method.
func (m *MockStatic) GetStatic(ctx context.Context, id int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStatic", ctx, id)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStatic indicates an expected call of GetStatic.
func (mr *MockStaticMockRecorder) GetStatic(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStatic", reflect.TypeOf((*MockStatic)(nil).GetStatic), ctx, id)
}

// UploadStatic mocks base method.
func (m *MockStatic) UploadStatic(ctx context.Context, data []byte) (*dto.UploadStaticResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadStatic", ctx, data)
	ret0, _ := ret[0].(*dto.UploadStaticResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadStatic indicates an expected call of UploadStatic.
func (mr *MockStaticMockRecorder) UploadStatic(ctx, data any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadStatic", reflect.TypeOf((*MockStatic)(nil).UploadStatic), ctx, data)
}
