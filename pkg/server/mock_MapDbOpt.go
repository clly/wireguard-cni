// Code generated by mockery v2.14.0. DO NOT EDIT.

package server

import mock "github.com/stretchr/testify/mock"

// MockMapDbOpt is an autogenerated mock type for the MapDbOpt type
type MockMapDbOpt struct {
	mock.Mock
}

// Execute provides a mock function with given fields: _a0
func (_m *MockMapDbOpt) Execute(_a0 *mapDB) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*mapDB) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockMapDbOpt interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockMapDbOpt creates a new instance of MockMapDbOpt. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockMapDbOpt(t mockConstructorTestingTNewMockMapDbOpt) *MockMapDbOpt {
	mock := &MockMapDbOpt{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
