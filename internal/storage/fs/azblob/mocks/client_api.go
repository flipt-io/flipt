// Code generated by mockery v2.36.1. DO NOT EDIT.

package mocks

import (
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	blob "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	container "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"

	context "context"

	mock "github.com/stretchr/testify/mock"

	runtime "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
)

// ClientAPI is an autogenerated mock type for the ClientAPI type
type ClientAPI struct {
	mock.Mock
}

// DownloadStream provides a mock function with given fields: ctx, containerName, blobName, o
func (_m *ClientAPI) DownloadStream(ctx context.Context, containerName string, blobName string, o *blob.DownloadStreamOptions) (blob.DownloadStreamResponse, error) {
	ret := _m.Called(ctx, containerName, blobName, o)

	var r0 blob.DownloadStreamResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *blob.DownloadStreamOptions) (blob.DownloadStreamResponse, error)); ok {
		return rf(ctx, containerName, blobName, o)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *blob.DownloadStreamOptions) blob.DownloadStreamResponse); ok {
		r0 = rf(ctx, containerName, blobName, o)
	} else {
		r0 = ret.Get(0).(blob.DownloadStreamResponse)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, *blob.DownloadStreamOptions) error); ok {
		r1 = rf(ctx, containerName, blobName, o)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewListBlobsFlatPager provides a mock function with given fields: containerName, o
func (_m *ClientAPI) NewListBlobsFlatPager(containerName string, o *container.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse] {
	ret := _m.Called(containerName, o)

	var r0 *runtime.Pager[azblob.ListBlobsFlatResponse]
	if rf, ok := ret.Get(0).(func(string, *container.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse]); ok {
		r0 = rf(containerName, o)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*runtime.Pager[azblob.ListBlobsFlatResponse])
		}
	}

	return r0
}

// NewClientAPI creates a new instance of ClientAPI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClientAPI(t interface {
	mock.TestingT
	Cleanup(func())
}) *ClientAPI {
	mock := &ClientAPI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}