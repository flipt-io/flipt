package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/storage"
	"go.uber.org/zap/zaptest"
)

func TestSetCacheHandleMarshalError(t *testing.T) {
	var (
		store          = &storeMock{}
		cacher         = &cacheSpy{}
		logger         = zaptest.NewLogger(t)
		cachedStore, _ = New(store, cacher, logger).(*Store)
	)

	cachedStore.setCache(context.TODO(), "key", make(chan int))
	assert.Empty(t, cacher.cacheKey)
}

func TestGetCacheHandleGetError(t *testing.T) {
	var (
		store          = &storeMock{}
		cacher         = &cacheSpy{getErr: errors.New("get error")}
		logger         = zaptest.NewLogger(t)
		cachedStore, _ = New(store, cacher, logger).(*Store)
	)

	value := make(map[string]string)
	cacheHit := cachedStore.getCache(context.TODO(), "key", &value)
	assert.False(t, cacheHit)
}

func TestGetCacheHandleUnmarshalError(t *testing.T) {
	var (
		store  = &storeMock{}
		cacher = &cacheSpy{
			cached:      true,
			cachedValue: []byte(`{"invalid":"123"`),
		}
		logger         = zaptest.NewLogger(t)
		cachedStore, _ = New(store, cacher, logger).(*Store)
	)

	value := make(map[string]string)
	cacheHit := cachedStore.getCache(context.TODO(), "key", &value)
	assert.False(t, cacheHit)
}

func TestGetEvaluationRules(t *testing.T) {
	var (
		expectedRules = []*storage.EvaluationRule{{ID: "123"}}
		store         = &storeMock{}
	)

	store.On("GetEvaluationRules", context.TODO(), "ns", "flag-1").Return(
		expectedRules, nil,
	)

	var (
		cacher      = &cacheSpy{}
		logger      = zaptest.NewLogger(t)
		cachedStore = New(store, cacher, logger)
	)

	rules, err := cachedStore.GetEvaluationRules(context.TODO(), "ns", "flag-1")
	assert.Nil(t, err)
	assert.Equal(t, expectedRules, rules)

	assert.Equal(t, "s:er:ns:flag-1", cacher.cacheKey)
	fmt.Println(string(cacher.cachedValue))
	assert.Equal(t, []byte(`[{"id":"123"}]`), cacher.cachedValue)
}

func TestGetEvaluationRulesCached(t *testing.T) {
	var (
		expectedRules = []*storage.EvaluationRule{{ID: "123"}}
		store         = &storeMock{}
	)

	store.On("GetEvaluationRules", context.TODO(), "ns", "flag-1").Return(
		expectedRules, nil,
	)

	var (
		cacher = &cacheSpy{
			cached:      true,
			cachedValue: []byte(`[{"id":"123"}]`),
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = New(store, cacher, logger)
	)

	rules, err := cachedStore.GetEvaluationRules(context.TODO(), "ns", "flag-1")
	assert.Nil(t, err)
	assert.Equal(t, expectedRules, rules)
	assert.Equal(t, "s:er:ns:flag-1", cacher.cacheKey)
}
