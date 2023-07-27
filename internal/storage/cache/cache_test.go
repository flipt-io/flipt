package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	"go.uber.org/zap/zaptest"
)

func TestSetCacheHandleMarshalError(t *testing.T) {
	store := &mockStore{}
	cacher := &mockCacher{}
	logger := zaptest.NewLogger(t)
	cachedStore, _ := New(store, cacher, logger).(*Store)

	cachedStore.setCache(context.TODO(), "key", make(chan int))
	require.Equal(t, "", cacher.CacheKey)
}

func TestGetCacheHandleGetError(t *testing.T) {
	store := &mockStore{}
	cacher := &mockCacher{GetErr: errors.New("get error")}
	logger := zaptest.NewLogger(t)
	cachedStore, _ := New(store, cacher, logger).(*Store)

	value := make(map[string]string)
	cacheHit := cachedStore.getCache(context.TODO(), "key", &value)
	require.False(t, cacheHit)
}

func TestGetCacheHandleUnmarshalError(t *testing.T) {
	store := &mockStore{}
	cacher := &mockCacher{
		Cached:      true,
		CachedValue: []byte(`{"invalid":"123"`),
	}
	logger := zaptest.NewLogger(t)
	cachedStore, _ := New(store, cacher, logger).(*Store)

	value := make(map[string]string)
	cacheHit := cachedStore.getCache(context.TODO(), "key", &value)
	require.False(t, cacheHit)
}

func TestGetEvaluationRules(t *testing.T) {
	expectedRules := []*storage.EvaluationRule{{ID: "123"}}
	store := &mockStore{
		EvaluationRules: expectedRules,
	}
	cacher := &mockCacher{}
	logger := zaptest.NewLogger(t)
	cachedStore := New(store, cacher, logger)

	rules, err := cachedStore.GetEvaluationRules(context.TODO(), "ns", "flag-1")
	require.Nil(t, err)
	require.Equal(t, expectedRules, rules)

	require.Equal(t, "s:er:ns:flag-1", cacher.CacheKey)
	require.Equal(t, []byte(`[{"id":"123"}]`), cacher.CachedValue)
}

func TestGetEvaluationRulesCached(t *testing.T) {
	expectedRules := []*storage.EvaluationRule{{ID: "123"}}
	store := &mockStore{
		EvaluationRules: []*storage.EvaluationRule{{ID: "543"}},
	}
	cacher := &mockCacher{
		Cached:      true,
		CachedValue: []byte(`[{"id":"123"}]`),
	}
	logger := zaptest.NewLogger(t)
	cachedStore := New(store, cacher, logger)

	rules, err := cachedStore.GetEvaluationRules(context.TODO(), "ns", "flag-1")
	require.Nil(t, err)
	require.Equal(t, expectedRules, rules)

	require.Equal(t, "s:er:ns:flag-1", cacher.CacheKey)
}
