package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	"go.uber.org/zap/zaptest"
)

func TestSetHandleMarshalError(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		cacher      = &cacheSpy{}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	cachedStore.setJSON(context.TODO(), "key", make(chan int))
	assert.Empty(t, cacher.cacheKey)
}

func TestGetHandleGetError(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		cacher      = &cacheSpy{getErr: errors.New("get error")}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	value := make(map[string]string)
	cacheHit := cachedStore.getJSON(context.TODO(), "key", &value)
	assert.False(t, cacheHit)
}

func TestGetHandleUnmarshalError(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		cacher = &cacheSpy{
			cached:      true,
			cachedValue: []byte(`{"invalid":"123"`),
		}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	value := make(map[string]string)
	cacheHit := cachedStore.getJSON(context.TODO(), "key", &value)
	assert.False(t, cacheHit)
}

func TestGetEvaluationRules(t *testing.T) {
	var (
		expectedRules = []*storage.EvaluationRule{{ID: "123"}}
		store         = &common.StoreMock{}
	)

	store.On("GetEvaluationRules", context.TODO(), storage.NewResource("ns", "flag-1")).Return(
		expectedRules, nil,
	)

	var (
		cacher      = &cacheSpy{}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	rules, err := cachedStore.GetEvaluationRules(context.TODO(), storage.NewResource("ns", "flag-1"))
	assert.Nil(t, err)
	assert.Equal(t, expectedRules, rules)

	assert.Equal(t, "s:er:ns:flag-1", cacher.cacheKey)
	assert.Equal(t, []byte(`[{"id":"123"}]`), cacher.cachedValue)
}

func TestGetEvaluationRulesCached(t *testing.T) {
	var (
		expectedRules = []*storage.EvaluationRule{{Rank: 12}}
		store         = &common.StoreMock{}
	)

	store.AssertNotCalled(t, "GetEvaluationRules", context.TODO(), storage.NewResource("ns", "flag-1"))

	var (
		cacher = &cacheSpy{
			cached:      true,
			cachedValue: []byte(`[{"rank":12}]`),
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	rules, err := cachedStore.GetEvaluationRules(context.TODO(), storage.NewResource("ns", "flag-1"))
	assert.Nil(t, err)
	assert.Equal(t, expectedRules, rules)
	assert.Equal(t, "s:er:ns:flag-1", cacher.cacheKey)
}

func TestGetEvaluationRollouts(t *testing.T) {
	var (
		expectedRollouts = []*storage.EvaluationRollout{{Rank: 1}}
		store            = &common.StoreMock{}
	)

	store.On("GetEvaluationRollouts", context.TODO(), storage.NewResource("ns", "flag-1")).Return(
		expectedRollouts, nil,
	)

	var (
		cacher      = &cacheSpy{}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	rollouts, err := cachedStore.GetEvaluationRollouts(context.TODO(), storage.NewResource("ns", "flag-1"))
	assert.Nil(t, err)
	assert.Equal(t, expectedRollouts, rollouts)

	assert.Equal(t, "s:ero:ns:flag-1", cacher.cacheKey)
	assert.Equal(t, []byte(`[{"rank":1}]`), cacher.cachedValue)
}

func TestGetEvaluationRolloutsCached(t *testing.T) {
	var (
		expectedRollouts = []*storage.EvaluationRollout{{Rank: 1}}
		store            = &common.StoreMock{}
	)

	store.AssertNotCalled(t, "GetEvaluationRollouts", context.TODO(), storage.NewResource("ns", "flag-1"))

	var (
		cacher = &cacheSpy{
			cached:      true,
			cachedValue: []byte(`[{"rank":1}]`),
		}

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	rollouts, err := cachedStore.GetEvaluationRollouts(context.TODO(), storage.NewResource("ns", "flag-1"))
	assert.Nil(t, err)
	assert.Equal(t, expectedRollouts, rollouts)
	assert.Equal(t, "s:ero:ns:flag-1", cacher.cacheKey)
}
