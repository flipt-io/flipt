package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/proto"
)

func TestGetAuthenticationByClientToken(t *testing.T) {
	var (
		auth  = &rpcauth.Authentication{Id: "123"}
		store = &storeMock{}
	)

	store.On("GetAuthenticationByClientToken", context.TODO(), "foo").Return(
		auth, nil,
	)

	var (
		cacher      = &cacheSpy{}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	got, err := cachedStore.GetAuthenticationByClientToken(context.TODO(), "foo")
	assert.Nil(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, auth.Id, got.Id)
	assert.Equal(t, "s:a:t:foo", cacher.cacheKey)
	assert.NotNil(t, cacher.cachedValue)
}

func TestGetEvaluationRulesCached(t *testing.T) {
	var (
		auth  = &rpcauth.Authentication{Id: "123"}
		store = &storeMock{}
	)

	store.AssertNotCalled(t, "GetAuthenticationByClientToken", context.TODO(), "foo")

	b, err := proto.Marshal(auth)
	require.NoError(t, err)

	var (
		cacher = &cacheSpy{
			cached:      true,
			cachedValue: b,
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	got, err := cachedStore.GetAuthenticationByClientToken(context.TODO(), "foo")
	assert.Nil(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, auth.Id, got.Id)
	assert.Equal(t, "s:a:t:foo", cacher.cacheKey)
}
