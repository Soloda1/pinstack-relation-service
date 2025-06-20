// Code generated by mockery v2.53.4. DO NOT EDIT.

package mocks

import (
	context "context"
	model "pinstack-relation-service/internal/model"

	mock "github.com/stretchr/testify/mock"
)

// FollowRepository is an autogenerated mock type for the FollowRepository type
type FollowRepository struct {
	mock.Mock
}

type FollowRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *FollowRepository) EXPECT() *FollowRepository_Expecter {
	return &FollowRepository_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: ctx, followerID, followeeID
func (_m *FollowRepository) Create(ctx context.Context, followerID int64, followeeID int64) (model.Follower, error) {
	ret := _m.Called(ctx, followerID, followeeID)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 model.Follower
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) (model.Follower, error)); ok {
		return rf(ctx, followerID, followeeID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) model.Follower); ok {
		r0 = rf(ctx, followerID, followeeID)
	} else {
		r0 = ret.Get(0).(model.Follower)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, int64) error); ok {
		r1 = rf(ctx, followerID, followeeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FollowRepository_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type FollowRepository_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - followerID int64
//   - followeeID int64
func (_e *FollowRepository_Expecter) Create(ctx interface{}, followerID interface{}, followeeID interface{}) *FollowRepository_Create_Call {
	return &FollowRepository_Create_Call{Call: _e.mock.On("Create", ctx, followerID, followeeID)}
}

func (_c *FollowRepository_Create_Call) Run(run func(ctx context.Context, followerID int64, followeeID int64)) *FollowRepository_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].(int64))
	})
	return _c
}

func (_c *FollowRepository_Create_Call) Return(_a0 model.Follower, _a1 error) *FollowRepository_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FollowRepository_Create_Call) RunAndReturn(run func(context.Context, int64, int64) (model.Follower, error)) *FollowRepository_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, followerID, followeeID
func (_m *FollowRepository) Delete(ctx context.Context, followerID int64, followeeID int64) error {
	ret := _m.Called(ctx, followerID, followeeID)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) error); ok {
		r0 = rf(ctx, followerID, followeeID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FollowRepository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type FollowRepository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - followerID int64
//   - followeeID int64
func (_e *FollowRepository_Expecter) Delete(ctx interface{}, followerID interface{}, followeeID interface{}) *FollowRepository_Delete_Call {
	return &FollowRepository_Delete_Call{Call: _e.mock.On("Delete", ctx, followerID, followeeID)}
}

func (_c *FollowRepository_Delete_Call) Run(run func(ctx context.Context, followerID int64, followeeID int64)) *FollowRepository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].(int64))
	})
	return _c
}

func (_c *FollowRepository_Delete_Call) Return(_a0 error) *FollowRepository_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *FollowRepository_Delete_Call) RunAndReturn(run func(context.Context, int64, int64) error) *FollowRepository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// Exists provides a mock function with given fields: ctx, followerID, followeeID
func (_m *FollowRepository) Exists(ctx context.Context, followerID int64, followeeID int64) (bool, error) {
	ret := _m.Called(ctx, followerID, followeeID)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) (bool, error)); ok {
		return rf(ctx, followerID, followeeID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, int64) bool); ok {
		r0 = rf(ctx, followerID, followeeID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, int64) error); ok {
		r1 = rf(ctx, followerID, followeeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FollowRepository_Exists_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exists'
type FollowRepository_Exists_Call struct {
	*mock.Call
}

// Exists is a helper method to define mock.On call
//   - ctx context.Context
//   - followerID int64
//   - followeeID int64
func (_e *FollowRepository_Expecter) Exists(ctx interface{}, followerID interface{}, followeeID interface{}) *FollowRepository_Exists_Call {
	return &FollowRepository_Exists_Call{Call: _e.mock.On("Exists", ctx, followerID, followeeID)}
}

func (_c *FollowRepository_Exists_Call) Run(run func(ctx context.Context, followerID int64, followeeID int64)) *FollowRepository_Exists_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].(int64))
	})
	return _c
}

func (_c *FollowRepository_Exists_Call) Return(_a0 bool, _a1 error) *FollowRepository_Exists_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FollowRepository_Exists_Call) RunAndReturn(run func(context.Context, int64, int64) (bool, error)) *FollowRepository_Exists_Call {
	_c.Call.Return(run)
	return _c
}

// GetFollowees provides a mock function with given fields: ctx, followerID, limit, offset
func (_m *FollowRepository) GetFollowees(ctx context.Context, followerID int64, limit int32, offset int32) ([]int64, error) {
	ret := _m.Called(ctx, followerID, limit, offset)

	if len(ret) == 0 {
		panic("no return value specified for GetFollowees")
	}

	var r0 []int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int32, int32) ([]int64, error)); ok {
		return rf(ctx, followerID, limit, offset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, int32, int32) []int64); ok {
		r0 = rf(ctx, followerID, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int64)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, int32, int32) error); ok {
		r1 = rf(ctx, followerID, limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FollowRepository_GetFollowees_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFollowees'
type FollowRepository_GetFollowees_Call struct {
	*mock.Call
}

// GetFollowees is a helper method to define mock.On call
//   - ctx context.Context
//   - followerID int64
//   - limit int32
//   - offset int32
func (_e *FollowRepository_Expecter) GetFollowees(ctx interface{}, followerID interface{}, limit interface{}, offset interface{}) *FollowRepository_GetFollowees_Call {
	return &FollowRepository_GetFollowees_Call{Call: _e.mock.On("GetFollowees", ctx, followerID, limit, offset)}
}

func (_c *FollowRepository_GetFollowees_Call) Run(run func(ctx context.Context, followerID int64, limit int32, offset int32)) *FollowRepository_GetFollowees_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].(int32), args[3].(int32))
	})
	return _c
}

func (_c *FollowRepository_GetFollowees_Call) Return(_a0 []int64, _a1 error) *FollowRepository_GetFollowees_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FollowRepository_GetFollowees_Call) RunAndReturn(run func(context.Context, int64, int32, int32) ([]int64, error)) *FollowRepository_GetFollowees_Call {
	_c.Call.Return(run)
	return _c
}

// GetFollowers provides a mock function with given fields: ctx, followeeID, limit, offset
func (_m *FollowRepository) GetFollowers(ctx context.Context, followeeID int64, limit int32, offset int32) ([]int64, error) {
	ret := _m.Called(ctx, followeeID, limit, offset)

	if len(ret) == 0 {
		panic("no return value specified for GetFollowers")
	}

	var r0 []int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, int32, int32) ([]int64, error)); ok {
		return rf(ctx, followeeID, limit, offset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int64, int32, int32) []int64); ok {
		r0 = rf(ctx, followeeID, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]int64)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int64, int32, int32) error); ok {
		r1 = rf(ctx, followeeID, limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FollowRepository_GetFollowers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFollowers'
type FollowRepository_GetFollowers_Call struct {
	*mock.Call
}

// GetFollowers is a helper method to define mock.On call
//   - ctx context.Context
//   - followeeID int64
//   - limit int32
//   - offset int32
func (_e *FollowRepository_Expecter) GetFollowers(ctx interface{}, followeeID interface{}, limit interface{}, offset interface{}) *FollowRepository_GetFollowers_Call {
	return &FollowRepository_GetFollowers_Call{Call: _e.mock.On("GetFollowers", ctx, followeeID, limit, offset)}
}

func (_c *FollowRepository_GetFollowers_Call) Run(run func(ctx context.Context, followeeID int64, limit int32, offset int32)) *FollowRepository_GetFollowers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].(int32), args[3].(int32))
	})
	return _c
}

func (_c *FollowRepository_GetFollowers_Call) Return(_a0 []int64, _a1 error) *FollowRepository_GetFollowers_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *FollowRepository_GetFollowers_Call) RunAndReturn(run func(context.Context, int64, int32, int32) ([]int64, error)) *FollowRepository_GetFollowers_Call {
	_c.Call.Return(run)
	return _c
}

// NewFollowRepository creates a new instance of FollowRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFollowRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *FollowRepository {
	mock := &FollowRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
