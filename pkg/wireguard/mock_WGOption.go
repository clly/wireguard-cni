// Code generated by mockery v2.40.1. DO NOT EDIT.

package wireguard

import mock "github.com/stretchr/testify/mock"

// MockWGOption is an autogenerated mock type for the WGOption type
type MockWGOption struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0
func (_m *MockWGOption) Execute(_a0 *WGQuickManager) {
	_m.Called(_a0)
}

// NewMockWGOption creates a new instance of MockWGOption. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockWGOption(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockWGOption {
	mock := &MockWGOption{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
