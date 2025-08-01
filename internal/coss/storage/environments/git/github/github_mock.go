// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package github

import (
	"context"

	"github.com/google/go-github/v66/github"
	mock "github.com/stretchr/testify/mock"
)

// NewMockPullRequestsService creates a new instance of MockPullRequestsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPullRequestsService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPullRequestsService {
	mock := &MockPullRequestsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockPullRequestsService is an autogenerated mock type for the PullRequestsService type
type MockPullRequestsService struct {
	mock.Mock
}

type MockPullRequestsService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPullRequestsService) EXPECT() *MockPullRequestsService_Expecter {
	return &MockPullRequestsService_Expecter{mock: &_m.Mock}
}

// Create provides a mock function for the type MockPullRequestsService
func (_mock *MockPullRequestsService) Create(ctx context.Context, owner string, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error) {
	ret := _mock.Called(ctx, owner, repo, pr)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *github.PullRequest
	var r1 *github.Response
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, *github.NewPullRequest) (*github.PullRequest, *github.Response, error)); ok {
		return returnFunc(ctx, owner, repo, pr)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, *github.NewPullRequest) *github.PullRequest); ok {
		r0 = returnFunc(ctx, owner, repo, pr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.PullRequest)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string, *github.NewPullRequest) *github.Response); ok {
		r1 = returnFunc(ctx, owner, repo, pr)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*github.Response)
		}
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, string, string, *github.NewPullRequest) error); ok {
		r2 = returnFunc(ctx, owner, repo, pr)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockPullRequestsService_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MockPullRequestsService_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx
//   - owner
//   - repo
//   - pr
func (_e *MockPullRequestsService_Expecter) Create(ctx interface{}, owner interface{}, repo interface{}, pr interface{}) *MockPullRequestsService_Create_Call {
	return &MockPullRequestsService_Create_Call{Call: _e.mock.On("Create", ctx, owner, repo, pr)}
}

func (_c *MockPullRequestsService_Create_Call) Run(run func(ctx context.Context, owner string, repo string, pr *github.NewPullRequest)) *MockPullRequestsService_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(*github.NewPullRequest))
	})
	return _c
}

func (_c *MockPullRequestsService_Create_Call) Return(pullRequest *github.PullRequest, response *github.Response, err error) *MockPullRequestsService_Create_Call {
	_c.Call.Return(pullRequest, response, err)
	return _c
}

func (_c *MockPullRequestsService_Create_Call) RunAndReturn(run func(ctx context.Context, owner string, repo string, pr *github.NewPullRequest) (*github.PullRequest, *github.Response, error)) *MockPullRequestsService_Create_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function for the type MockPullRequestsService
func (_mock *MockPullRequestsService) List(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error) {
	ret := _mock.Called(ctx, owner, repo, opts)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []*github.PullRequest
	var r1 *github.Response
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)); ok {
		return returnFunc(ctx, owner, repo, opts)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, *github.PullRequestListOptions) []*github.PullRequest); ok {
		r0 = returnFunc(ctx, owner, repo, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.PullRequest)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string, *github.PullRequestListOptions) *github.Response); ok {
		r1 = returnFunc(ctx, owner, repo, opts)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*github.Response)
		}
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, string, string, *github.PullRequestListOptions) error); ok {
		r2 = returnFunc(ctx, owner, repo, opts)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockPullRequestsService_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type MockPullRequestsService_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx
//   - owner
//   - repo
//   - opts
func (_e *MockPullRequestsService_Expecter) List(ctx interface{}, owner interface{}, repo interface{}, opts interface{}) *MockPullRequestsService_List_Call {
	return &MockPullRequestsService_List_Call{Call: _e.mock.On("List", ctx, owner, repo, opts)}
}

func (_c *MockPullRequestsService_List_Call) Run(run func(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions)) *MockPullRequestsService_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(*github.PullRequestListOptions))
	})
	return _c
}

func (_c *MockPullRequestsService_List_Call) Return(pullRequests []*github.PullRequest, response *github.Response, err error) *MockPullRequestsService_List_Call {
	_c.Call.Return(pullRequests, response, err)
	return _c
}

func (_c *MockPullRequestsService_List_Call) RunAndReturn(run func(ctx context.Context, owner string, repo string, opts *github.PullRequestListOptions) ([]*github.PullRequest, *github.Response, error)) *MockPullRequestsService_List_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockRepositoriesService creates a new instance of MockRepositoriesService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockRepositoriesService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockRepositoriesService {
	mock := &MockRepositoriesService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockRepositoriesService is an autogenerated mock type for the RepositoriesService type
type MockRepositoriesService struct {
	mock.Mock
}

type MockRepositoriesService_Expecter struct {
	mock *mock.Mock
}

func (_m *MockRepositoriesService) EXPECT() *MockRepositoriesService_Expecter {
	return &MockRepositoriesService_Expecter{mock: &_m.Mock}
}

// CompareCommits provides a mock function for the type MockRepositoriesService
func (_mock *MockRepositoriesService) CompareCommits(ctx context.Context, owner string, repo string, base string, head string, opts *github.ListOptions) (*github.CommitsComparison, *github.Response, error) {
	ret := _mock.Called(ctx, owner, repo, base, head, opts)

	if len(ret) == 0 {
		panic("no return value specified for CompareCommits")
	}

	var r0 *github.CommitsComparison
	var r1 *github.Response
	var r2 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, string, string, *github.ListOptions) (*github.CommitsComparison, *github.Response, error)); ok {
		return returnFunc(ctx, owner, repo, base, head, opts)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string, string, string, *github.ListOptions) *github.CommitsComparison); ok {
		r0 = returnFunc(ctx, owner, repo, base, head, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.CommitsComparison)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string, string, string, *github.ListOptions) *github.Response); ok {
		r1 = returnFunc(ctx, owner, repo, base, head, opts)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*github.Response)
		}
	}
	if returnFunc, ok := ret.Get(2).(func(context.Context, string, string, string, string, *github.ListOptions) error); ok {
		r2 = returnFunc(ctx, owner, repo, base, head, opts)
	} else {
		r2 = ret.Error(2)
	}
	return r0, r1, r2
}

// MockRepositoriesService_CompareCommits_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompareCommits'
type MockRepositoriesService_CompareCommits_Call struct {
	*mock.Call
}

// CompareCommits is a helper method to define mock.On call
//   - ctx
//   - owner
//   - repo
//   - base
//   - head
//   - opts
func (_e *MockRepositoriesService_Expecter) CompareCommits(ctx interface{}, owner interface{}, repo interface{}, base interface{}, head interface{}, opts interface{}) *MockRepositoriesService_CompareCommits_Call {
	return &MockRepositoriesService_CompareCommits_Call{Call: _e.mock.On("CompareCommits", ctx, owner, repo, base, head, opts)}
}

func (_c *MockRepositoriesService_CompareCommits_Call) Run(run func(ctx context.Context, owner string, repo string, base string, head string, opts *github.ListOptions)) *MockRepositoriesService_CompareCommits_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(string), args[4].(string), args[5].(*github.ListOptions))
	})
	return _c
}

func (_c *MockRepositoriesService_CompareCommits_Call) Return(commitsComparison *github.CommitsComparison, response *github.Response, err error) *MockRepositoriesService_CompareCommits_Call {
	_c.Call.Return(commitsComparison, response, err)
	return _c
}

func (_c *MockRepositoriesService_CompareCommits_Call) RunAndReturn(run func(ctx context.Context, owner string, repo string, base string, head string, opts *github.ListOptions) (*github.CommitsComparison, *github.Response, error)) *MockRepositoriesService_CompareCommits_Call {
	_c.Call.Return(run)
	return _c
}
