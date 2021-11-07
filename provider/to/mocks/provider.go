// Code generated by MockGen. DO NOT EDIT.
// Source: provider.go

// Package mock_toprovider is a generated GoMock package.
package mock_toprovider

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	toprovider "github.com/grezar/revolver/provider/to"
)

// MockProvider is a mock of Provider interface.
type MockProvider struct {
	ctrl     *gomock.Controller
	recorder *MockProviderMockRecorder
}

// MockProviderMockRecorder is the mock recorder for MockProvider.
type MockProviderMockRecorder struct {
	mock *MockProvider
}

// NewMockProvider creates a new mock instance.
func NewMockProvider(ctrl *gomock.Controller) *MockProvider {
	mock := &MockProvider{ctrl: ctrl}
	mock.recorder = &MockProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockProvider) EXPECT() *MockProviderMockRecorder {
	return m.recorder
}

// Name mocks base method.
func (m *MockProvider) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockProviderMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockProvider)(nil).Name))
}

// UnmarshalSpec mocks base method.
func (m *MockProvider) UnmarshalSpec(bytes []byte) (toprovider.Operator, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnmarshalSpec", bytes)
	ret0, _ := ret[0].(toprovider.Operator)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UnmarshalSpec indicates an expected call of UnmarshalSpec.
func (mr *MockProviderMockRecorder) UnmarshalSpec(bytes interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnmarshalSpec", reflect.TypeOf((*MockProvider)(nil).UnmarshalSpec), bytes)
}

// MockOperator is a mock of Operator interface.
type MockOperator struct {
	ctrl     *gomock.Controller
	recorder *MockOperatorMockRecorder
}

// MockOperatorMockRecorder is the mock recorder for MockOperator.
type MockOperatorMockRecorder struct {
	mock *MockOperator
}

// NewMockOperator creates a new mock instance.
func NewMockOperator(ctrl *gomock.Controller) *MockOperator {
	mock := &MockOperator{ctrl: ctrl}
	mock.recorder = &MockOperatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockOperator) EXPECT() *MockOperatorMockRecorder {
	return m.recorder
}

// UpdateSecret mocks base method.
func (m *MockOperator) UpdateSecret(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateSecret", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateSecret indicates an expected call of UpdateSecret.
func (mr *MockOperatorMockRecorder) UpdateSecret(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateSecret", reflect.TypeOf((*MockOperator)(nil).UpdateSecret), ctx)
}
