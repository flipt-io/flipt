// Code generated by mockery v2.53.3. DO NOT EDIT.

package environments

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
	evaluation "go.flipt.io/flipt/rpc/flipt/evaluation"

	storage "go.flipt.io/flipt/internal/storage"

	v2environments "go.flipt.io/flipt/rpc/v2/environments"
)

// MockEnvironment is an autogenerated mock type for the Environment type
type MockEnvironment struct {
	mock.Mock
}

type MockEnvironment_Expecter struct {
	mock *mock.Mock
}

func (_m *MockEnvironment) EXPECT() *MockEnvironment_Expecter {
	return &MockEnvironment_Expecter{mock: &_m.Mock}
}

// CreateNamespace provides a mock function with given fields: _a0, rev, _a2
func (_m *MockEnvironment) CreateNamespace(_a0 context.Context, rev string, _a2 *v2environments.Namespace) (string, error) {
	ret := _m.Called(_a0, rev, _a2)

	if len(ret) == 0 {
		panic("no return value specified for CreateNamespace")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *v2environments.Namespace) (string, error)); ok {
		return rf(_a0, rev, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *v2environments.Namespace) string); ok {
		r0 = rf(_a0, rev, _a2)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *v2environments.Namespace) error); ok {
		r1 = rf(_a0, rev, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_CreateNamespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateNamespace'
type MockEnvironment_CreateNamespace_Call struct {
	*mock.Call
}

// CreateNamespace is a helper method to define mock.On call
//   - _a0 context.Context
//   - rev string
//   - _a2 *v2environments.Namespace
func (_e *MockEnvironment_Expecter) CreateNamespace(_a0 interface{}, rev interface{}, _a2 interface{}) *MockEnvironment_CreateNamespace_Call {
	return &MockEnvironment_CreateNamespace_Call{Call: _e.mock.On("CreateNamespace", _a0, rev, _a2)}
}

func (_c *MockEnvironment_CreateNamespace_Call) Run(run func(_a0 context.Context, rev string, _a2 *v2environments.Namespace)) *MockEnvironment_CreateNamespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*v2environments.Namespace))
	})
	return _c
}

func (_c *MockEnvironment_CreateNamespace_Call) Return(_a0 string, _a1 error) *MockEnvironment_CreateNamespace_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_CreateNamespace_Call) RunAndReturn(run func(context.Context, string, *v2environments.Namespace) (string, error)) *MockEnvironment_CreateNamespace_Call {
	_c.Call.Return(run)
	return _c
}

// Default provides a mock function with no fields
func (_m *MockEnvironment) Default() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Default")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockEnvironment_Default_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Default'
type MockEnvironment_Default_Call struct {
	*mock.Call
}

// Default is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) Default() *MockEnvironment_Default_Call {
	return &MockEnvironment_Default_Call{Call: _e.mock.On("Default")}
}

func (_c *MockEnvironment_Default_Call) Run(run func()) *MockEnvironment_Default_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_Default_Call) Return(_a0 bool) *MockEnvironment_Default_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_Default_Call) RunAndReturn(run func() bool) *MockEnvironment_Default_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteNamespace provides a mock function with given fields: _a0, rev, key
func (_m *MockEnvironment) DeleteNamespace(_a0 context.Context, rev string, key string) (string, error) {
	ret := _m.Called(_a0, rev, key)

	if len(ret) == 0 {
		panic("no return value specified for DeleteNamespace")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (string, error)); ok {
		return rf(_a0, rev, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(_a0, rev, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(_a0, rev, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_DeleteNamespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteNamespace'
type MockEnvironment_DeleteNamespace_Call struct {
	*mock.Call
}

// DeleteNamespace is a helper method to define mock.On call
//   - _a0 context.Context
//   - rev string
//   - key string
func (_e *MockEnvironment_Expecter) DeleteNamespace(_a0 interface{}, rev interface{}, key interface{}) *MockEnvironment_DeleteNamespace_Call {
	return &MockEnvironment_DeleteNamespace_Call{Call: _e.mock.On("DeleteNamespace", _a0, rev, key)}
}

func (_c *MockEnvironment_DeleteNamespace_Call) Run(run func(_a0 context.Context, rev string, key string)) *MockEnvironment_DeleteNamespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockEnvironment_DeleteNamespace_Call) Return(_a0 string, _a1 error) *MockEnvironment_DeleteNamespace_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_DeleteNamespace_Call) RunAndReturn(run func(context.Context, string, string) (string, error)) *MockEnvironment_DeleteNamespace_Call {
	_c.Call.Return(run)
	return _c
}

// EvaluationNamespaceSnapshot provides a mock function with given fields: _a0, _a1
func (_m *MockEnvironment) EvaluationNamespaceSnapshot(_a0 context.Context, _a1 string) (*evaluation.EvaluationNamespaceSnapshot, error) {
	ret := _m.Called(_a0, _a1)

	if len(ret) == 0 {
		panic("no return value specified for EvaluationNamespaceSnapshot")
	}

	var r0 *evaluation.EvaluationNamespaceSnapshot
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*evaluation.EvaluationNamespaceSnapshot, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *evaluation.EvaluationNamespaceSnapshot); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evaluation.EvaluationNamespaceSnapshot)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_EvaluationNamespaceSnapshot_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EvaluationNamespaceSnapshot'
type MockEnvironment_EvaluationNamespaceSnapshot_Call struct {
	*mock.Call
}

// EvaluationNamespaceSnapshot is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 string
func (_e *MockEnvironment_Expecter) EvaluationNamespaceSnapshot(_a0 interface{}, _a1 interface{}) *MockEnvironment_EvaluationNamespaceSnapshot_Call {
	return &MockEnvironment_EvaluationNamespaceSnapshot_Call{Call: _e.mock.On("EvaluationNamespaceSnapshot", _a0, _a1)}
}

func (_c *MockEnvironment_EvaluationNamespaceSnapshot_Call) Run(run func(_a0 context.Context, _a1 string)) *MockEnvironment_EvaluationNamespaceSnapshot_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockEnvironment_EvaluationNamespaceSnapshot_Call) Return(_a0 *evaluation.EvaluationNamespaceSnapshot, _a1 error) *MockEnvironment_EvaluationNamespaceSnapshot_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_EvaluationNamespaceSnapshot_Call) RunAndReturn(run func(context.Context, string) (*evaluation.EvaluationNamespaceSnapshot, error)) *MockEnvironment_EvaluationNamespaceSnapshot_Call {
	_c.Call.Return(run)
	return _c
}

// EvaluationStore provides a mock function with no fields
func (_m *MockEnvironment) EvaluationStore() (storage.ReadOnlyStore, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EvaluationStore")
	}

	var r0 storage.ReadOnlyStore
	var r1 error
	if rf, ok := ret.Get(0).(func() (storage.ReadOnlyStore, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() storage.ReadOnlyStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storage.ReadOnlyStore)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_EvaluationStore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EvaluationStore'
type MockEnvironment_EvaluationStore_Call struct {
	*mock.Call
}

// EvaluationStore is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) EvaluationStore() *MockEnvironment_EvaluationStore_Call {
	return &MockEnvironment_EvaluationStore_Call{Call: _e.mock.On("EvaluationStore")}
}

func (_c *MockEnvironment_EvaluationStore_Call) Run(run func()) *MockEnvironment_EvaluationStore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_EvaluationStore_Call) Return(_a0 storage.ReadOnlyStore, _a1 error) *MockEnvironment_EvaluationStore_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_EvaluationStore_Call) RunAndReturn(run func() (storage.ReadOnlyStore, error)) *MockEnvironment_EvaluationStore_Call {
	_c.Call.Return(run)
	return _c
}

// GetNamespace provides a mock function with given fields: _a0, key
func (_m *MockEnvironment) GetNamespace(_a0 context.Context, key string) (*v2environments.NamespaceResponse, error) {
	ret := _m.Called(_a0, key)

	if len(ret) == 0 {
		panic("no return value specified for GetNamespace")
	}

	var r0 *v2environments.NamespaceResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*v2environments.NamespaceResponse, error)); ok {
		return rf(_a0, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *v2environments.NamespaceResponse); ok {
		r0 = rf(_a0, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v2environments.NamespaceResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(_a0, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_GetNamespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNamespace'
type MockEnvironment_GetNamespace_Call struct {
	*mock.Call
}

// GetNamespace is a helper method to define mock.On call
//   - _a0 context.Context
//   - key string
func (_e *MockEnvironment_Expecter) GetNamespace(_a0 interface{}, key interface{}) *MockEnvironment_GetNamespace_Call {
	return &MockEnvironment_GetNamespace_Call{Call: _e.mock.On("GetNamespace", _a0, key)}
}

func (_c *MockEnvironment_GetNamespace_Call) Run(run func(_a0 context.Context, key string)) *MockEnvironment_GetNamespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockEnvironment_GetNamespace_Call) Return(_a0 *v2environments.NamespaceResponse, _a1 error) *MockEnvironment_GetNamespace_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_GetNamespace_Call) RunAndReturn(run func(context.Context, string) (*v2environments.NamespaceResponse, error)) *MockEnvironment_GetNamespace_Call {
	_c.Call.Return(run)
	return _c
}

// Key provides a mock function with no fields
func (_m *MockEnvironment) Key() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Key")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockEnvironment_Key_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Key'
type MockEnvironment_Key_Call struct {
	*mock.Call
}

// Key is a helper method to define mock.On call
func (_e *MockEnvironment_Expecter) Key() *MockEnvironment_Key_Call {
	return &MockEnvironment_Key_Call{Call: _e.mock.On("Key")}
}

func (_c *MockEnvironment_Key_Call) Run(run func()) *MockEnvironment_Key_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockEnvironment_Key_Call) Return(_a0 string) *MockEnvironment_Key_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_Key_Call) RunAndReturn(run func() string) *MockEnvironment_Key_Call {
	_c.Call.Return(run)
	return _c
}

// ListNamespaces provides a mock function with given fields: _a0
func (_m *MockEnvironment) ListNamespaces(_a0 context.Context) (*v2environments.ListNamespacesResponse, error) {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for ListNamespaces")
	}

	var r0 *v2environments.ListNamespacesResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (*v2environments.ListNamespacesResponse, error)); ok {
		return rf(_a0)
	}
	if rf, ok := ret.Get(0).(func(context.Context) *v2environments.ListNamespacesResponse); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v2environments.ListNamespacesResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_ListNamespaces_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListNamespaces'
type MockEnvironment_ListNamespaces_Call struct {
	*mock.Call
}

// ListNamespaces is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *MockEnvironment_Expecter) ListNamespaces(_a0 interface{}) *MockEnvironment_ListNamespaces_Call {
	return &MockEnvironment_ListNamespaces_Call{Call: _e.mock.On("ListNamespaces", _a0)}
}

func (_c *MockEnvironment_ListNamespaces_Call) Run(run func(_a0 context.Context)) *MockEnvironment_ListNamespaces_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockEnvironment_ListNamespaces_Call) Return(_a0 *v2environments.ListNamespacesResponse, _a1 error) *MockEnvironment_ListNamespaces_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_ListNamespaces_Call) RunAndReturn(run func(context.Context) (*v2environments.ListNamespacesResponse, error)) *MockEnvironment_ListNamespaces_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: _a0, rev, typ, fn
func (_m *MockEnvironment) Update(_a0 context.Context, rev string, typ ResourceType, fn UpdateFunc) (string, error) {
	ret := _m.Called(_a0, rev, typ, fn)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ResourceType, UpdateFunc) (string, error)); ok {
		return rf(_a0, rev, typ, fn)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ResourceType, UpdateFunc) string); ok {
		r0 = rf(_a0, rev, typ, fn)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ResourceType, UpdateFunc) error); ok {
		r1 = rf(_a0, rev, typ, fn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type MockEnvironment_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - _a0 context.Context
//   - rev string
//   - typ ResourceType
//   - fn UpdateFunc
func (_e *MockEnvironment_Expecter) Update(_a0 interface{}, rev interface{}, typ interface{}, fn interface{}) *MockEnvironment_Update_Call {
	return &MockEnvironment_Update_Call{Call: _e.mock.On("Update", _a0, rev, typ, fn)}
}

func (_c *MockEnvironment_Update_Call) Run(run func(_a0 context.Context, rev string, typ ResourceType, fn UpdateFunc)) *MockEnvironment_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(ResourceType), args[3].(UpdateFunc))
	})
	return _c
}

func (_c *MockEnvironment_Update_Call) Return(_a0 string, _a1 error) *MockEnvironment_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_Update_Call) RunAndReturn(run func(context.Context, string, ResourceType, UpdateFunc) (string, error)) *MockEnvironment_Update_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateNamespace provides a mock function with given fields: _a0, rev, _a2
func (_m *MockEnvironment) UpdateNamespace(_a0 context.Context, rev string, _a2 *v2environments.Namespace) (string, error) {
	ret := _m.Called(_a0, rev, _a2)

	if len(ret) == 0 {
		panic("no return value specified for UpdateNamespace")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *v2environments.Namespace) (string, error)); ok {
		return rf(_a0, rev, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *v2environments.Namespace) string); ok {
		r0 = rf(_a0, rev, _a2)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *v2environments.Namespace) error); ok {
		r1 = rf(_a0, rev, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockEnvironment_UpdateNamespace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateNamespace'
type MockEnvironment_UpdateNamespace_Call struct {
	*mock.Call
}

// UpdateNamespace is a helper method to define mock.On call
//   - _a0 context.Context
//   - rev string
//   - _a2 *v2environments.Namespace
func (_e *MockEnvironment_Expecter) UpdateNamespace(_a0 interface{}, rev interface{}, _a2 interface{}) *MockEnvironment_UpdateNamespace_Call {
	return &MockEnvironment_UpdateNamespace_Call{Call: _e.mock.On("UpdateNamespace", _a0, rev, _a2)}
}

func (_c *MockEnvironment_UpdateNamespace_Call) Run(run func(_a0 context.Context, rev string, _a2 *v2environments.Namespace)) *MockEnvironment_UpdateNamespace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*v2environments.Namespace))
	})
	return _c
}

func (_c *MockEnvironment_UpdateNamespace_Call) Return(_a0 string, _a1 error) *MockEnvironment_UpdateNamespace_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockEnvironment_UpdateNamespace_Call) RunAndReturn(run func(context.Context, string, *v2environments.Namespace) (string, error)) *MockEnvironment_UpdateNamespace_Call {
	_c.Call.Return(run)
	return _c
}

// View provides a mock function with given fields: _a0, typ, fn
func (_m *MockEnvironment) View(_a0 context.Context, typ ResourceType, fn ViewFunc) error {
	ret := _m.Called(_a0, typ, fn)

	if len(ret) == 0 {
		panic("no return value specified for View")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ResourceType, ViewFunc) error); ok {
		r0 = rf(_a0, typ, fn)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockEnvironment_View_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'View'
type MockEnvironment_View_Call struct {
	*mock.Call
}

// View is a helper method to define mock.On call
//   - _a0 context.Context
//   - typ ResourceType
//   - fn ViewFunc
func (_e *MockEnvironment_Expecter) View(_a0 interface{}, typ interface{}, fn interface{}) *MockEnvironment_View_Call {
	return &MockEnvironment_View_Call{Call: _e.mock.On("View", _a0, typ, fn)}
}

func (_c *MockEnvironment_View_Call) Run(run func(_a0 context.Context, typ ResourceType, fn ViewFunc)) *MockEnvironment_View_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(ResourceType), args[2].(ViewFunc))
	})
	return _c
}

func (_c *MockEnvironment_View_Call) Return(_a0 error) *MockEnvironment_View_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockEnvironment_View_Call) RunAndReturn(run func(context.Context, ResourceType, ViewFunc) error) *MockEnvironment_View_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockEnvironment creates a new instance of MockEnvironment. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockEnvironment(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockEnvironment {
	mock := &MockEnvironment{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
