package cache

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/auth"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ cache.Cacher = &cacheSpy{}

type cacheSpy struct {
	getCacheKeys    []string
	setCacheKeys    []string
	deleteCacheKeys []string

	cache map[string][]byte

	getErr      error
	setErr      error
	deleteError error
}

func (c *cacheSpy) String() string {
	return "mockCacher"
}

func (c *cacheSpy) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.getCacheKeys = append(c.getCacheKeys, key)

	if c.getErr != nil {
		return nil, false, c.getErr
	}

	got, ok := c.cache[key]
	return got, ok, nil
}

func (c *cacheSpy) Set(ctx context.Context, key string, value []byte) error {
	c.setCacheKeys = append(c.setCacheKeys, key)

	if c.setErr != nil {
		return c.setErr
	}

	if c.cache == nil {
		c.cache = make(map[string][]byte)
	}

	c.cache[key] = value
	return nil
}

func (c *cacheSpy) Delete(ctx context.Context, key string) error {
	c.deleteCacheKeys = append(c.deleteCacheKeys, key)

	if c.deleteError != nil {
		return c.deleteError
	}

	delete(c.cache, key)
	return nil
}

var _ auth.Store = &storeMock{}

type storeMock struct {
	mock.Mock
}

func (m *storeMock) String() string {
	return "mock"
}

func (m *storeMock) CreateAuthentication(ctx context.Context, r *auth.CreateAuthenticationRequest) (string, *rpcauth.Authentication, error) {
	args := m.Called(ctx, r)
	return args.String(0), args.Get(1).(*rpcauth.Authentication), args.Error(2)
}

func (m *storeMock) GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*rpcauth.Authentication, error) {
	args := m.Called(ctx, clientToken)
	return args.Get(0).(*rpcauth.Authentication), args.Error(1)
}

func (m *storeMock) GetAuthenticationByID(ctx context.Context, id string) (*rpcauth.Authentication, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*rpcauth.Authentication), args.Error(1)
}

func (m *storeMock) ListAuthentications(ctx context.Context, r *storage.ListRequest[auth.ListAuthenticationsPredicate]) (storage.ResultSet[*rpcauth.Authentication], error) {
	args := m.Called(ctx, r)
	return args.Get(0).(storage.ResultSet[*rpcauth.Authentication]), args.Error(1)

}

func (m *storeMock) DeleteAuthentications(ctx context.Context, r *auth.DeleteAuthenticationsRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) ExpireAuthenticationByID(ctx context.Context, id string, ts *timestamppb.Timestamp) error {
	args := m.Called(ctx, id, ts)
	return args.Error(0)
}
