// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Hasher is an autogenerated mock type for the Hasher type
type Hasher struct {
	mock.Mock
}

// Compare provides a mock function with given fields: ctx, hash, token
func (_m *Hasher) Compare(ctx context.Context, hash string, token string) error {
	ret := _m.Called(ctx, hash, token)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, hash, token)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Generate provides a mock function with given fields: ctx, token
func (_m *Hasher) Generate(ctx context.Context, token string) (string, error) {
	ret := _m.Called(ctx, token)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, token)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, token)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewHasher creates a new instance of Hasher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHasher(t interface {
	mock.TestingT
	Cleanup(func())
}) *Hasher {
	mock := &Hasher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
