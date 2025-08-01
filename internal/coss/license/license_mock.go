// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package license

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/product"
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

// Product provides a mock function for the type MockManager
func (_mock *MockManager) Product() product.Product {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Product")
	}

	var r0 product.Product
	if returnFunc, ok := ret.Get(0).(func() product.Product); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(product.Product)
	}
	return r0
}

// MockManager_Product_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Product'
type MockManager_Product_Call struct {
	*mock.Call
}

// Product is a helper method to define mock.On call
func (_e *MockManager_Expecter) Product() *MockManager_Product_Call {
	return &MockManager_Product_Call{Call: _e.mock.On("Product")}
}

func (_c *MockManager_Product_Call) Run(run func()) *MockManager_Product_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_Product_Call) Return(product1 product.Product) *MockManager_Product_Call {
	_c.Call.Return(product1)
	return _c
}

func (_c *MockManager_Product_Call) RunAndReturn(run func() product.Product) *MockManager_Product_Call {
	_c.Call.Return(run)
	return _c
}

// Shutdown provides a mock function for the type MockManager
func (_mock *MockManager) Shutdown(ctx context.Context) error {
	ret := _mock.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Shutdown")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = returnFunc(ctx)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// MockManager_Shutdown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Shutdown'
type MockManager_Shutdown_Call struct {
	*mock.Call
}

// Shutdown is a helper method to define mock.On call
//   - ctx
func (_e *MockManager_Expecter) Shutdown(ctx interface{}) *MockManager_Shutdown_Call {
	return &MockManager_Shutdown_Call{Call: _e.mock.On("Shutdown", ctx)}
}

func (_c *MockManager_Shutdown_Call) Run(run func(ctx context.Context)) *MockManager_Shutdown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockManager_Shutdown_Call) Return(err error) *MockManager_Shutdown_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *MockManager_Shutdown_Call) RunAndReturn(run func(ctx context.Context) error) *MockManager_Shutdown_Call {
	_c.Call.Return(run)
	return _c
}
