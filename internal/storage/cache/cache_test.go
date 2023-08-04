package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/storage"
	"go.uber.org/zap/zaptest"
)

func TestSetHandleMarshalError(t *testing.T) {
	var (
		store       = &storeMock{}
		cacher      = &cacheSpy{}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	cachedStore.set(context.TODO(), "key", make(chan int))
	assert.Empty(t, cacher.cacheKey)
}

func TestGetHandleGetError(t *testing.T) {
	var (
		store       = &storeMock{}
		cacher      = &cacheSpy{getErr: errors.New("get error")}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	value := make(map[string]string)
	cacheHit := cachedStore.get(context.TODO(), "key", &value)
	assert.False(t, cacheHit)
}

func TestGetHandleUnmarshalError(t *testing.T) {
	var (
		store  = &storeMock{}
		cacher = &cacheSpy{
			cached:      true,
			cachedValue: []byte(`{"invalid":"123"`),
		}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	value := make(map[string]string)
	cacheHit := cachedStore.get(context.TODO(), "key", &value)
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
		cachedStore = NewStore(store, cacher, logger)
	)

	rules, err := cachedStore.GetEvaluationRules(context.TODO(), "ns", "flag-1")
	assert.Nil(t, err)
	assert.Equal(t, expectedRules, rules)

	assert.Equal(t, "s:er:ns:flag-1", cacher.cacheKey)
	assert.Equal(t, []byte(`[{"id":"123"}]`), cacher.cachedValue)
}

func TestGetEvaluationRulesCached(t *testing.T) {
	var (
		expectedRules = []*storage.EvaluationRule{{ID: "123"}}
		store         = &storeMock{}
	)

	store.AssertNotCalled(t, "GetEvaluationRules", context.TODO(), "ns", "flag-1")

	var (
		cacher = &cacheSpy{
			cached:      true,
			cachedValue: []byte(`[{"id":"123"}]`),
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	rules, err := cachedStore.GetEvaluationRules(context.TODO(), "ns", "flag-1")
	assert.Nil(t, err)
	assert.Equal(t, expectedRules, rules)
	assert.Equal(t, "s:er:ns:flag-1", cacher.cacheKey)
}
