// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package secrets

import (
	"context"

	mock "github.com/stretchr/testify/mock"
)

// NewMockManager creates a new instance of MockManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockManager {
	mock := &MockManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockManager is an autogenerated mock type for the Manager type
type MockManager struct {
	mock.Mock
}

type MockManager_Expecter struct {
	mock *mock.Mock
}

func (_m *MockManager) EXPECT() *MockManager_Expecter {
	return &MockManager_Expecter{mock: &_m.Mock}
}

// Close provides a mock function for the type MockManager
func (_mock *MockManager) Close() error {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func() error); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockManager_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockManager_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockManager_Expecter) Close() *MockManager_Close_Call {
	return &MockManager_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockManager_Close_Call) Run(run func()) *MockManager_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_Close_Call) Return(err error) *MockManager_Close_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockManager_Close_Call) RunAndReturn(run func() error) *MockManager_Close_Call {
	_c.Call.Return(run)
	return _c
}

// GetProvider provides a mock function for the type MockManager
func (_mock *MockManager) GetProvider(name string) (Provider, error) {
	ret := _mock.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for GetProvider")
	}

	var r0 Provider
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(string) (Provider, error)); ok {
		return returnFunc(name)
	}
	if returnFunc, ok := ret.Get(0).(func(string) Provider); ok {
		r0 = returnFunc(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Provider)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(string) error); ok {
		r1 = returnFunc(name)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockManager_GetProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetProvider'
type MockManager_GetProvider_Call struct {
	*mock.Call
}

// GetProvider is a helper method to define mock.On call
//   - name
func (_e *MockManager_Expecter) GetProvider(name interface{}) *MockManager_GetProvider_Call {
	return &MockManager_GetProvider_Call{Call: _e.mock.On("GetProvider", name)}
}

func (_c *MockManager_GetProvider_Call) Run(run func(name string)) *MockManager_GetProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockManager_GetProvider_Call) Return(provider Provider, err error) *MockManager_GetProvider_Call {
	_c.Call.Return(provider, err)
	return _c
}

func (_c *MockManager_GetProvider_Call) RunAndReturn(run func(name string) (Provider, error)) *MockManager_GetProvider_Call {
	_c.Call.Return(run)
	return _c
}

// GetSecret provides a mock function for the type MockManager
func (_mock *MockManager) GetSecret(ctx context.Context, providerName string, path string) (*Secret, error) {
	ret := _mock.Called(ctx, providerName, path)

	if len(ret) == 0 {
		panic("no return value specified for GetSecret")
	}

	var r0 *Secret
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) (*Secret, error)); ok {
		return returnFunc(ctx, providerName, path)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) *Secret); ok {
		r0 = returnFunc(ctx, providerName, path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Secret)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = returnFunc(ctx, providerName, path)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockManager_GetSecret_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSecret'
type MockManager_GetSecret_Call struct {
	*mock.Call
}

// GetSecret is a helper method to define mock.On call
//   - ctx
//   - providerName
//   - path
func (_e *MockManager_Expecter) GetSecret(ctx interface{}, providerName interface{}, path interface{}) *MockManager_GetSecret_Call {
	return &MockManager_GetSecret_Call{Call: _e.mock.On("GetSecret", ctx, providerName, path)}
}

func (_c *MockManager_GetSecret_Call) Run(run func(ctx context.Context, providerName string, path string)) *MockManager_GetSecret_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockManager_GetSecret_Call) Return(secret *Secret, err error) *MockManager_GetSecret_Call {
	_c.Call.Return(secret, err)
	return _c
}

func (_c *MockManager_GetSecret_Call) RunAndReturn(run func(ctx context.Context, providerName string, path string) (*Secret, error)) *MockManager_GetSecret_Call {
	_c.Call.Return(run)
	return _c
}

// GetSecretValue provides a mock function for the type MockManager
func (_mock *MockManager) GetSecretValue(ctx context.Context, ref Reference) ([]byte, error) {
	ret := _mock.Called(ctx, ref)

	if len(ret) == 0 {
		panic("no return value specified for GetSecretValue")
	}

	var r0 []byte
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, Reference) ([]byte, error)); ok {
		return returnFunc(ctx, ref)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, Reference) []byte); ok {
		r0 = returnFunc(ctx, ref)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, Reference) error); ok {
		r1 = returnFunc(ctx, ref)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockManager_GetSecretValue_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSecretValue'
type MockManager_GetSecretValue_Call struct {
	*mock.Call
}

// GetSecretValue is a helper method to define mock.On call
//   - ctx
//   - ref
func (_e *MockManager_Expecter) GetSecretValue(ctx interface{}, ref interface{}) *MockManager_GetSecretValue_Call {
	return &MockManager_GetSecretValue_Call{Call: _e.mock.On("GetSecretValue", ctx, ref)}
}

func (_c *MockManager_GetSecretValue_Call) Run(run func(ctx context.Context, ref Reference)) *MockManager_GetSecretValue_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(Reference))
	})
	return _c
}

func (_c *MockManager_GetSecretValue_Call) Return(bytes []byte, err error) *MockManager_GetSecretValue_Call {
	_c.Call.Return(bytes, err)
	return _c
}

func (_c *MockManager_GetSecretValue_Call) RunAndReturn(run func(ctx context.Context, ref Reference) ([]byte, error)) *MockManager_GetSecretValue_Call {
	_c.Call.Return(run)
	return _c
}

// ListProviders provides a mock function for the type MockManager
func (_mock *MockManager) ListProviders() []string {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for ListProviders")
	}

	var r0 []string
	if returnFunc, ok := ret.Get(0).(func() []string); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}
	return r0
}

// MockManager_ListProviders_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListProviders'
type MockManager_ListProviders_Call struct {
	*mock.Call
}

// ListProviders is a helper method to define mock.On call
func (_e *MockManager_Expecter) ListProviders() *MockManager_ListProviders_Call {
	return &MockManager_ListProviders_Call{Call: _e.mock.On("ListProviders")}
}

func (_c *MockManager_ListProviders_Call) Run(run func()) *MockManager_ListProviders_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_ListProviders_Call) Return(strings []string) *MockManager_ListProviders_Call {
	_c.Call.Return(strings)
	return _c
}

func (_c *MockManager_ListProviders_Call) RunAndReturn(run func() []string) *MockManager_ListProviders_Call {
	_c.Call.Return(run)
	return _c
}

// ListSecrets provides a mock function for the type MockManager
func (_mock *MockManager) ListSecrets(ctx context.Context, providerName string, pathPrefix string) ([]string, error) {
	ret := _mock.Called(ctx, providerName, pathPrefix)

	if len(ret) == 0 {
		panic("no return value specified for ListSecrets")
	}

	var r0 []string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) ([]string, error)); ok {
		return returnFunc(ctx, providerName, pathPrefix)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) []string); ok {
		r0 = returnFunc(ctx, providerName, pathPrefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = returnFunc(ctx, providerName, pathPrefix)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockManager_ListSecrets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListSecrets'
type MockManager_ListSecrets_Call struct {
	*mock.Call
}

// ListSecrets is a helper method to define mock.On call
//   - ctx
//   - providerName
//   - pathPrefix
func (_e *MockManager_Expecter) ListSecrets(ctx interface{}, providerName interface{}, pathPrefix interface{}) *MockManager_ListSecrets_Call {
	return &MockManager_ListSecrets_Call{Call: _e.mock.On("ListSecrets", ctx, providerName, pathPrefix)}
}

func (_c *MockManager_ListSecrets_Call) Run(run func(ctx context.Context, providerName string, pathPrefix string)) *MockManager_ListSecrets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockManager_ListSecrets_Call) Return(strings []string, err error) *MockManager_ListSecrets_Call {
	_c.Call.Return(strings, err)
	return _c
}

func (_c *MockManager_ListSecrets_Call) RunAndReturn(run func(ctx context.Context, providerName string, pathPrefix string) ([]string, error)) *MockManager_ListSecrets_Call {
	_c.Call.Return(run)
	return _c
}

// RegisterProvider provides a mock function for the type MockManager
func (_mock *MockManager) RegisterProvider(name string, provider Provider) error {
	ret := _mock.Called(name, provider)

	if len(ret) == 0 {
		panic("no return value specified for RegisterProvider")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(string, Provider) error); ok {
		r0 = returnFunc(name, provider)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockManager_RegisterProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RegisterProvider'
type MockManager_RegisterProvider_Call struct {
	*mock.Call
}

// RegisterProvider is a helper method to define mock.On call
//   - name
//   - provider
func (_e *MockManager_Expecter) RegisterProvider(name interface{}, provider interface{}) *MockManager_RegisterProvider_Call {
	return &MockManager_RegisterProvider_Call{Call: _e.mock.On("RegisterProvider", name, provider)}
}

func (_c *MockManager_RegisterProvider_Call) Run(run func(name string, provider Provider)) *MockManager_RegisterProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(Provider))
	})
	return _c
}

func (_c *MockManager_RegisterProvider_Call) Return(err error) *MockManager_RegisterProvider_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockManager_RegisterProvider_Call) RunAndReturn(run func(name string, provider Provider) error) *MockManager_RegisterProvider_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockProvider creates a new instance of MockProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockProvider {
	mock := &MockProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockProvider is an autogenerated mock type for the Provider type
type MockProvider struct {
	mock.Mock
}

type MockProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *MockProvider) EXPECT() *MockProvider_Expecter {
	return &MockProvider_Expecter{mock: &_m.Mock}
}

// GetSecret provides a mock function for the type MockProvider
func (_mock *MockProvider) GetSecret(ctx context.Context, path string) (*Secret, error) {
	ret := _mock.Called(ctx, path)

	if len(ret) == 0 {
		panic("no return value specified for GetSecret")
	}

	var r0 *Secret
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (*Secret, error)); ok {
		return returnFunc(ctx, path)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) *Secret); ok {
		r0 = returnFunc(ctx, path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Secret)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, path)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockProvider_GetSecret_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSecret'
type MockProvider_GetSecret_Call struct {
	*mock.Call
}

// GetSecret is a helper method to define mock.On call
//   - ctx
//   - path
func (_e *MockProvider_Expecter) GetSecret(ctx interface{}, path interface{}) *MockProvider_GetSecret_Call {
	return &MockProvider_GetSecret_Call{Call: _e.mock.On("GetSecret", ctx, path)}
}

func (_c *MockProvider_GetSecret_Call) Run(run func(ctx context.Context, path string)) *MockProvider_GetSecret_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProvider_GetSecret_Call) Return(secret *Secret, err error) *MockProvider_GetSecret_Call {
	_c.Call.Return(secret, err)
	return _c
}

func (_c *MockProvider_GetSecret_Call) RunAndReturn(run func(ctx context.Context, path string) (*Secret, error)) *MockProvider_GetSecret_Call {
	_c.Call.Return(run)
	return _c
}

// ListSecrets provides a mock function for the type MockProvider
func (_mock *MockProvider) ListSecrets(ctx context.Context, pathPrefix string) ([]string, error) {
	ret := _mock.Called(ctx, pathPrefix)

	if len(ret) == 0 {
		panic("no return value specified for ListSecrets")
	}

	var r0 []string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) ([]string, error)); ok {
		return returnFunc(ctx, pathPrefix)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) []string); ok {
		r0 = returnFunc(ctx, pathPrefix)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, pathPrefix)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockProvider_ListSecrets_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ListSecrets'
type MockProvider_ListSecrets_Call struct {
	*mock.Call
}

// ListSecrets is a helper method to define mock.On call
//   - ctx
//   - pathPrefix
func (_e *MockProvider_Expecter) ListSecrets(ctx interface{}, pathPrefix interface{}) *MockProvider_ListSecrets_Call {
	return &MockProvider_ListSecrets_Call{Call: _e.mock.On("ListSecrets", ctx, pathPrefix)}
}

func (_c *MockProvider_ListSecrets_Call) Run(run func(ctx context.Context, pathPrefix string)) *MockProvider_ListSecrets_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockProvider_ListSecrets_Call) Return(strings []string, err error) *MockProvider_ListSecrets_Call {
	_c.Call.Return(strings, err)
	return _c
}

func (_c *MockProvider_ListSecrets_Call) RunAndReturn(run func(ctx context.Context, pathPrefix string) ([]string, error)) *MockProvider_ListSecrets_Call {
	_c.Call.Return(run)
	return _c
}
