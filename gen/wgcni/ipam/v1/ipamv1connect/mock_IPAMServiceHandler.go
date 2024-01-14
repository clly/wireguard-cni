// Code generated by mockery v2.40.1. DO NOT EDIT.

package ipamv1connect

import (
	context "context"

	connect "connectrpc.com/connect"

	ipamv1 "github.com/clly/wireguard-cni/gen/wgcni/ipam/v1"

	mock "github.com/stretchr/testify/mock"
)

// MockIPAMServiceHandler is an autogenerated mock type for the IPAMServiceHandler type
type MockIPAMServiceHandler struct {
	mock.Mock
}

// Alloc provides a mock function with given fields: _a0, _a1
func (_m *MockIPAMServiceHandler) Alloc(_a0 context.Context, _a1 *connect.Request[ipamv1.AllocRequest]) (*connect.Response[ipamv1.AllocResponse], error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Alloc")
	}

	var r0 *connect.Response[ipamv1.AllocResponse]
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *connect.Request[ipamv1.AllocRequest]) (*connect.Response[ipamv1.AllocResponse], error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *connect.Request[ipamv1.AllocRequest]) *connect.Response[ipamv1.AllocResponse]); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*connect.Response[ipamv1.AllocResponse])
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *connect.Request[ipamv1.AllocRequest]) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMockIPAMServiceHandler creates a new instance of MockIPAMServiceHandler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockIPAMServiceHandler(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIPAMServiceHandler {
	mock := &MockIPAMServiceHandler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
