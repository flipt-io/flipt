package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/proto"
)

//nolint:gosec
const hashedToken = "LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564="

func TestGetAuthenticationByClientToken(t *testing.T) {
	var (
		auth  = &rpcauth.Authentication{Id: "123"}
		store = &storeMock{
			StoreMock: &common.StoreMock{},
		}
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
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, auth.Id, got.Id)
	assert.Equal(t, fmt.Sprintf("s:a:t:%s", hashedToken), cacher.getCacheKeys[0]) // assert we looked up the key first
	assert.Equal(t, fmt.Sprintf("s:a:t:%s", hashedToken), cacher.setCacheKeys[0]) // assert we cached the key
	assert.Equal(t, "s:a:i:123", cacher.setCacheKeys[1])                          // assert we cached the id
	assert.Equal(t, []byte(hashedToken), cacher.cache["s:a:i:123"])               // assert we cached the token
	assert.NotNil(t, cacher.cache[fmt.Sprintf("s:a:t:%s", hashedToken)])          // assert we cached the auth
}

func TestGetAuthenticationByClientTokenCached(t *testing.T) {
	var (
		auth  = &rpcauth.Authentication{Id: "123"}
		store = &storeMock{
			StoreMock: &common.StoreMock{},
		}
	)

	store.AssertNotCalled(t, "GetAuthenticationByClientToken", mock.Anything, mock.Anything)

	b, err := proto.Marshal(auth)
	require.NoError(t, err)

	var (
		cacher = &cacheSpy{
			cache: map[string][]byte{
				fmt.Sprintf("s:a:t:%s", hashedToken): b,
			},
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	got, err := cachedStore.GetAuthenticationByClientToken(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, auth.Id, got.Id)
	assert.Equal(t, fmt.Sprintf("s:a:t:%s", hashedToken), cacher.getCacheKeys[0]) // assert we looked up the key
}

func TestExpireAuthenticationByID(t *testing.T) {
	store := &storeMock{
		StoreMock: &common.StoreMock{},
	}

	store.On("ExpireAuthenticationByID", context.TODO(), "123", mock.Anything).Return(
		nil,
	)

	var (
		cacher = &cacheSpy{
			cache: map[string][]byte{
				"s:a:i:123":                          []byte(hashedToken),
				fmt.Sprintf("s:a:t:%s", hashedToken): []byte("deleteme"),
			},
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	err := cachedStore.ExpireAuthenticationByID(context.TODO(), "123", nil)
	require.NoError(t, err)

	assert.Equal(t, "s:a:i:123", cacher.getCacheKeys[0])                             // assert we looked up the key by id
	assert.Equal(t, fmt.Sprintf("s:a:t:%s", hashedToken), cacher.deleteCacheKeys[0]) // assert we cached the key
	assert.Equal(t, "s:a:i:123", cacher.deleteCacheKeys[1])                          // assert we cached the id
}
