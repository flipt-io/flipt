package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	assert.Equal(t, "s:a:t:foo", cacher.getCacheKeys[0])      // assert we looked up the key first
	assert.Equal(t, "s:a:t:foo", cacher.setCacheKeys[0])      // assert we cached the key
	assert.Equal(t, "s:a:i:123", cacher.setCacheKeys[1])      // assert we cached the id
	assert.Equal(t, []byte("foo"), cacher.cache["s:a:i:123"]) // assert we cached the token
	assert.NotNil(t, cacher.cache["s:a:t:foo"])               // assert we cached the auth
}

func TestGetAuthenticationByClientTokenCached(t *testing.T) {
	var (
		auth  = &rpcauth.Authentication{Id: "123"}
		store = &storeMock{}
	)

	store.AssertNotCalled(t, "GetAuthenticationByClientToken", mock.Anything, mock.Anything)

	b, err := proto.Marshal(auth)
	require.NoError(t, err)

	var (
		cacher = &cacheSpy{
			cache: map[string][]byte{
				"s:a:t:foo": b,
			},
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	got, err := cachedStore.GetAuthenticationByClientToken(context.TODO(), "foo")
	assert.Nil(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, auth.Id, got.Id)
	assert.Equal(t, "s:a:t:foo", cacher.getCacheKeys[0]) // assert we looked up the key
}

func TestExpireAuthenticationByID(t *testing.T) {
	store := &storeMock{}

	store.On("ExpireAuthenticationByID", context.TODO(), "123", mock.Anything).Return(
		nil,
	)

	var (
		cacher = &cacheSpy{
			cache: map[string][]byte{
				"s:a:i:123": []byte("foo"),
				"s:a:t:foo": []byte("deleteme"),
			},
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	err := cachedStore.ExpireAuthenticationByID(context.TODO(), "123", nil)
	assert.Nil(t, err)

	assert.Equal(t, "s:a:i:123", cacher.getCacheKeys[0])    // assert we looked up the key by id
	assert.Equal(t, "s:a:t:foo", cacher.deleteCacheKeys[0]) // assert we cached the key
	assert.Equal(t, "s:a:i:123", cacher.deleteCacheKeys[1]) // assert we cached the id
}
