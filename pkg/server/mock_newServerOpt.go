// Code generated by mockery v2.40.1. DO NOT EDIT.

package server

import mock "github.com/stretchr/testify/mock"

// mockNewServerOpt is an autogenerated mock type for the newServerOpt type
type mockNewServerOpt struct {
	mock.Mock
}

// Execute provides a mock function with given fields: cfg
func (_m *mockNewServerOpt) Execute(cfg *serverConfig) {
	_m.Called(cfg)
}

// newMockNewServerOpt creates a new instance of mockNewServerOpt. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockNewServerOpt(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockNewServerOpt {
	mock := &mockNewServerOpt{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
