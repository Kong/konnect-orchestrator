// Code generated by mockery. DO NOT EDIT.

package team

import (
	context "context"

	components "github.com/Kong/sdk-konnect-go/models/components"

	mock "github.com/stretchr/testify/mock"

	operations "github.com/Kong/sdk-konnect-go/models/operations"
)

// MockTeamService is an autogenerated mock type for the TeamService type
type MockTeamService struct {
	mock.Mock
}

type MockTeamService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTeamService) EXPECT() *MockTeamService_Expecter {
	return &MockTeamService_Expecter{mock: &_m.Mock}
}

// CreateTeam provides a mock function with given fields: ctx, request, opts
func (_m *MockTeamService) CreateTeam(ctx context.Context, request *components.CreateTeam, opts ...operations.Option) (*operations.CreateTeamResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateTeam")
	}

	var r0 *operations.CreateTeamResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *components.CreateTeam, ...operations.Option) (*operations.CreateTeamResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *components.CreateTeam, ...operations.Option) *operations.CreateTeamResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.CreateTeamResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *components.CreateTeam, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTeamService_CreateTeam_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateTeam'
type MockTeamService_CreateTeam_Call struct {
	*mock.Call
}

// CreateTeam is a helper method to define mock.On call
//   - ctx context.Context
//   - request *components.CreateTeam
//   - opts ...operations.Option
func (_e *MockTeamService_Expecter) CreateTeam(ctx interface{}, request interface{}, opts ...interface{}) *MockTeamService_CreateTeam_Call {
	return &MockTeamService_CreateTeam_Call{Call: _e.mock.On("CreateTeam",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockTeamService_CreateTeam_Call) Run(run func(ctx context.Context, request *components.CreateTeam, opts ...operations.Option)) *MockTeamService_CreateTeam_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(*components.CreateTeam), variadicArgs...)
	})
	return _c
}

func (_c *MockTeamService_CreateTeam_Call) Return(_a0 *operations.CreateTeamResponse, _a1 error) *MockTeamService_CreateTeam_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTeamService_CreateTeam_Call) RunAndReturn(run func(context.Context, *components.CreateTeam, ...operations.Option) (*operations.CreateTeamResponse, error)) *MockTeamService_CreateTeam_Call {
	_c.Call.Return(run)
	return _c
}

// GetTeam provides a mock function with given fields: ctx, teamID, opts
func (_m *MockTeamService) GetTeam(ctx context.Context, teamID string, opts ...operations.Option) (*operations.GetTeamResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, teamID)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for GetTeam")
	}

	var r0 *operations.GetTeamResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...operations.Option) (*operations.GetTeamResponse, error)); ok {
		return rf(ctx, teamID, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...operations.Option) *operations.GetTeamResponse); ok {
		r0 = rf(ctx, teamID, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.GetTeamResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...operations.Option) error); ok {
		r1 = rf(ctx, teamID, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTeamService_GetTeam_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTeam'
type MockTeamService_GetTeam_Call struct {
	*mock.Call
}

// GetTeam is a helper method to define mock.On call
//   - ctx context.Context
//   - teamID string
//   - opts ...operations.Option
func (_e *MockTeamService_Expecter) GetTeam(ctx interface{}, teamID interface{}, opts ...interface{}) *MockTeamService_GetTeam_Call {
	return &MockTeamService_GetTeam_Call{Call: _e.mock.On("GetTeam",
		append([]interface{}{ctx, teamID}, opts...)...)}
}

func (_c *MockTeamService_GetTeam_Call) Run(run func(ctx context.Context, teamID string, opts ...operations.Option)) *MockTeamService_GetTeam_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), variadicArgs...)
	})
	return _c
}

func (_c *MockTeamService_GetTeam_Call) Return(_a0 *operations.GetTeamResponse, _a1 error) *MockTeamService_GetTeam_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTeamService_GetTeam_Call) RunAndReturn(run func(context.Context, string, ...operations.Option) (*operations.GetTeamResponse, error)) *MockTeamService_GetTeam_Call {
	_c.Call.Return(run)
	return _c
}

// ListTeams provides a mock function with given fields: ctx, request, opts
func (_m *MockTeamService) ListTeams(ctx context.Context, request operations.ListTeamsRequest, opts ...operations.Option) (*operations.ListTeamsResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for ListTeams")
	}

	var r0 *operations.ListTeamsResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListTeamsRequest, ...operations.Option) (*operations.ListTeamsResponse, error)); ok {
		return rf(ctx, request, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, operations.ListTeamsRequest, ...operations.Option) *operations.ListTeamsResponse); ok {
		r0 = rf(ctx, request, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.ListTeamsResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, operations.ListTeamsRequest, ...operations.Option) error); ok {
		r1 = rf(ctx, request, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTeamService_ListTeams_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListTeams'
type MockTeamService_ListTeams_Call struct {
	*mock.Call
}

// ListTeams is a helper method to define mock.On call
//   - ctx context.Context
//   - request operations.ListTeamsRequest
//   - opts ...operations.Option
func (_e *MockTeamService_Expecter) ListTeams(ctx interface{}, request interface{}, opts ...interface{}) *MockTeamService_ListTeams_Call {
	return &MockTeamService_ListTeams_Call{Call: _e.mock.On("ListTeams",
		append([]interface{}{ctx, request}, opts...)...)}
}

func (_c *MockTeamService_ListTeams_Call) Run(run func(ctx context.Context, request operations.ListTeamsRequest, opts ...operations.Option)) *MockTeamService_ListTeams_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(operations.ListTeamsRequest), variadicArgs...)
	})
	return _c
}

func (_c *MockTeamService_ListTeams_Call) Return(_a0 *operations.ListTeamsResponse, _a1 error) *MockTeamService_ListTeams_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTeamService_ListTeams_Call) RunAndReturn(run func(context.Context, operations.ListTeamsRequest, ...operations.Option) (*operations.ListTeamsResponse, error)) *MockTeamService_ListTeams_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateTeam provides a mock function with given fields: ctx, teamID, updateTeam, opts
func (_m *MockTeamService) UpdateTeam(ctx context.Context, teamID string, updateTeam *components.UpdateTeam, opts ...operations.Option) (*operations.UpdateTeamResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, teamID, updateTeam)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for UpdateTeam")
	}

	var r0 *operations.UpdateTeamResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *components.UpdateTeam, ...operations.Option) (*operations.UpdateTeamResponse, error)); ok {
		return rf(ctx, teamID, updateTeam, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *components.UpdateTeam, ...operations.Option) *operations.UpdateTeamResponse); ok {
		r0 = rf(ctx, teamID, updateTeam, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operations.UpdateTeamResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *components.UpdateTeam, ...operations.Option) error); ok {
		r1 = rf(ctx, teamID, updateTeam, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTeamService_UpdateTeam_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateTeam'
type MockTeamService_UpdateTeam_Call struct {
	*mock.Call
}

// UpdateTeam is a helper method to define mock.On call
//   - ctx context.Context
//   - teamID string
//   - updateTeam *components.UpdateTeam
//   - opts ...operations.Option
func (_e *MockTeamService_Expecter) UpdateTeam(ctx interface{}, teamID interface{}, updateTeam interface{}, opts ...interface{}) *MockTeamService_UpdateTeam_Call {
	return &MockTeamService_UpdateTeam_Call{Call: _e.mock.On("UpdateTeam",
		append([]interface{}{ctx, teamID, updateTeam}, opts...)...)}
}

func (_c *MockTeamService_UpdateTeam_Call) Run(run func(ctx context.Context, teamID string, updateTeam *components.UpdateTeam, opts ...operations.Option)) *MockTeamService_UpdateTeam_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]operations.Option, len(args)-3)
		for i, a := range args[3:] {
			if a != nil {
				variadicArgs[i] = a.(operations.Option)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(*components.UpdateTeam), variadicArgs...)
	})
	return _c
}

func (_c *MockTeamService_UpdateTeam_Call) Return(_a0 *operations.UpdateTeamResponse, _a1 error) *MockTeamService_UpdateTeam_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockTeamService_UpdateTeam_Call) RunAndReturn(run func(context.Context, string, *components.UpdateTeam, ...operations.Option) (*operations.UpdateTeamResponse, error)) *MockTeamService_UpdateTeam_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockTeamService creates a new instance of MockTeamService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTeamService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTeamService {
	mock := &MockTeamService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
