package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockUserJWTBuilder is a mock of UserJWTBuilder interface.
type MockUserJWTBuilder struct {
	ctrl     *gomock.Controller
	recorder *MockUserJWTBuilderMockRecorder
}

// MockUserJWTBuilderMockRecorder is the mock recorder for MockUserJWTBuilder.
type MockUserJWTBuilderMockRecorder struct {
	mock *MockUserJWTBuilder
}

// NewMockUserJWTBuilder creates a new mock instance.
func NewMockUserJWTBuilder(ctrl *gomock.Controller) *MockUserJWTBuilder {
	mock := &MockUserJWTBuilder{ctrl: ctrl}
	mock.recorder = &MockUserJWTBuilderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserJWTBuilder) EXPECT() *MockUserJWTBuilderMockRecorder {
	return m.recorder
}

// BuildJWTString mocks base method.
func (m *MockUserJWTBuilder) BuildJWTString(userID int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildJWTString", userID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildJWTString indicates an expected call of BuildJWTString.
func (mr *MockUserJWTBuilderMockRecorder) BuildJWTString(userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildJWTString", reflect.TypeOf((*MockUserJWTBuilder)(nil).BuildJWTString), userID)
}
