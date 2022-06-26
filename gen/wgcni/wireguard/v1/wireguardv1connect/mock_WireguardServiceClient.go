// Code generated by mockery v2.13.1. DO NOT EDIT.

package wireguardv1connect

import (
	context "context"

	connect "github.com/bufbuild/connect-go"

	mock "github.com/stretchr/testify/mock"

	wireguardv1 "wireguard-cni/gen/wgcni/wireguard/v1"
)

// MockWireguardServiceClient is an autogenerated mock type for the WireguardServiceClient type
type MockWireguardServiceClient struct {
	mock.Mock
}

// Peers provides a mock function with given fields: _a0, _a1
func (_m *MockWireguardServiceClient) Peers(_a0 context.Context, _a1 *connect.Request[wireguardv1.PeersRequest]) (*connect.Response[wireguardv1.PeersResponse], error) {
	ret := _m.Called(_a0, _a1)

	var r0 *connect.Response[wireguardv1.PeersResponse]
	if rf, ok := ret.Get(0).(func(context.Context, *connect.Request[wireguardv1.PeersRequest]) *connect.Response[wireguardv1.PeersResponse]); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*connect.Response[wireguardv1.PeersResponse])
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *connect.Request[wireguardv1.PeersRequest]) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Register provides a mock function with given fields: _a0, _a1
func (_m *MockWireguardServiceClient) Register(_a0 context.Context, _a1 *connect.Request[wireguardv1.RegisterRequest]) (*connect.Response[wireguardv1.RegisterResponse], error) {
	ret := _m.Called(_a0, _a1)

	var r0 *connect.Response[wireguardv1.RegisterResponse]
	if rf, ok := ret.Get(0).(func(context.Context, *connect.Request[wireguardv1.RegisterRequest]) *connect.Response[wireguardv1.RegisterResponse]); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*connect.Response[wireguardv1.RegisterResponse])
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *connect.Request[wireguardv1.RegisterRequest]) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewMockWireguardServiceClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockWireguardServiceClient creates a new instance of MockWireguardServiceClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockWireguardServiceClient(t mockConstructorTestingTNewMockWireguardServiceClient) *MockWireguardServiceClient {
	mock := &MockWireguardServiceClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}