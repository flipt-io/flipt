package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/cache/memory"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func TestSetJSON_HandleMarshalError(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		cacher      = &cacheMock{}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	cachedStore.setJSON(context.TODO(), "key", make(chan int))
	assert.Empty(t, cacher.cacheKey)
}

func TestGetJSON_HandleGetError(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		cacher      = &cacheMock{getErr: errors.New("get error")}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	value := make(map[string]string)
	cacheHit := cachedStore.getJSON(context.TODO(), "key", &value)
	assert.False(t, cacheHit)
}

func TestGetJSON_HandleUnmarshalError(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		cacher = &cacheMock{
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

func TestGetProtobuf_HandleGetError(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		cacher      = &cacheMock{getErr: errors.New("get error")}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	value := &flipt.Flag{}
	cacheHit := cachedStore.getProtobuf(context.TODO(), "key", value)
	assert.False(t, cacheHit)
}

func TestGetProtobuf_HandleUnmarshalError(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		cacher = &cacheMock{
			cached:      true,
			cachedValue: []byte(`{"invalid":"123"`),
		}
		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	value := &flipt.Flag{}
	cacheHit := cachedStore.getProtobuf(context.TODO(), "key", value)
	assert.False(t, cacheHit)
}

func TestGetEvaluationRules(t *testing.T) {
	var (
		expectedRules = []*storage.EvaluationRule{{NamespaceKey: "ns", Rank: 1}}
		store         = &common.StoreMock{}
	)

	store.On("GetVersion", context.TODO(), storage.NewNamespace("ns")).Return(
		"v-123", nil,
	)

	store.On("GetEvaluationRules", context.TODO(), storage.NewResource("ns", "flag-1")).Return(
		expectedRules, nil,
	)

	var (
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheBackendMemory,
		})
		cacher = newCacheSpy(cache)

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	// First call to get rules should call the store and cache the result
	_, err := cachedStore.GetEvaluationRules(context.TODO(), storage.NewResource("ns", "flag-1"))
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.setItems)
	assert.NotEmpty(t, cacher.setItems["s:er:ns:v-123:flag-1"])
	assert.NotEmpty(t, cacher.setItems["s:n:ns"])
	assert.Equal(t, 2, cacher.setCalled)

	// Second call to get rules should hit the cache
	_, err = cachedStore.GetEvaluationRules(context.TODO(), storage.NewResource("ns", "flag-1"))
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.getKeys)
	_, ok := cacher.getKeys["s:er:ns:v-123:flag-1"]
	assert.True(t, ok)

	_, ok = cacher.getKeys["s:n:ns"]
	assert.True(t, ok)

	store.AssertNumberOfCalls(t, "GetEvaluationRules", 1)
}

func TestGetEvaluationRollouts(t *testing.T) {
	var (
		expectedRollouts = []*storage.EvaluationRollout{{NamespaceKey: "ns", Rank: 1}}
		store            = &common.StoreMock{}
	)

	store.On("GetVersion", context.TODO(), storage.NewNamespace("ns")).Return(
		"v-321", nil,
	)

	store.On("GetEvaluationRollouts", context.TODO(), storage.NewResource("ns", "flag-1")).Return(
		expectedRollouts, nil,
	)

	var (
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheBackendMemory,
		})
		cacher = newCacheSpy(cache)

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	// First call to get rollouts should call the store and cache the result
	_, err := cachedStore.GetEvaluationRollouts(context.TODO(), storage.NewResource("ns", "flag-1"))
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.setItems)
	assert.NotEmpty(t, cacher.setItems["s:ero:ns:v-321:flag-1"])
	assert.NotEmpty(t, cacher.setItems["s:n:ns"])
	assert.Equal(t, 2, cacher.setCalled)

	// Second call to get rollouts should hit the cache
	_, err = cachedStore.GetEvaluationRollouts(context.TODO(), storage.NewResource("ns", "flag-1"))
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.getKeys)
	_, ok := cacher.getKeys["s:ero:ns:v-321:flag-1"]
	assert.True(t, ok)

	store.AssertNumberOfCalls(t, "GetEvaluationRollouts", 1)
}

func TestGetEvaluationDistributions(t *testing.T) {
	var (
		expectedDistributions = []*storage.EvaluationDistribution{{VariantKey: "v"}}
		store                 = &common.StoreMock{}
	)

	store.On("GetVersion", context.TODO(), storage.NewNamespace("ns")).Return(
		"v-321", nil,
	)

	store.On("GetEvaluationDistributions", context.TODO(), mock.Anything).Return(
		expectedDistributions, nil,
	)

	var (
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheBackendMemory,
		})
		cacher = newCacheSpy(cache)

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	// First call to get rollouts should call the store and cache the result
	_, err := cachedStore.GetEvaluationDistributions(context.TODO(), storage.NewResource("ns", "flag-1"), storage.IDRequest{ID: "v"})
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.setItems)
	assert.NotEmpty(t, cacher.setItems["s:erd:ns:v-321:flag-1/v"])
	assert.NotEmpty(t, cacher.setItems["s:n:ns"])
	assert.Equal(t, 2, cacher.setCalled)

	// Second call to get rollouts should hit the cache
	_, err = cachedStore.GetEvaluationDistributions(context.TODO(), storage.NewResource("ns", "flag-1"), storage.IDRequest{ID: "v"})
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.getKeys)
	_, ok := cacher.getKeys["s:erd:ns:v-321:flag-1/v"]
	assert.True(t, ok)

	store.AssertNumberOfCalls(t, "GetEvaluationDistributions", 1)
}

func TestGetFlag(t *testing.T) {
	var (
		expectedFlag = &flipt.Flag{NamespaceKey: "ns", Key: "flag-1"}
		store        = &common.StoreMock{}
	)

	store.On("GetVersion", context.TODO(), storage.NewNamespace("ns")).Return(
		"v-321", nil,
	)
	store.On("GetFlag", context.TODO(), storage.NewResource("ns", "flag-1")).Return(
		expectedFlag, nil,
	)

	var (
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheBackendMemory,
		})
		cacher = newCacheSpy(cache)

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	// First call to get flag should call the store and cache the result
	_, err := cachedStore.GetFlag(context.TODO(), storage.NewResource("ns", "flag-1"))
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.setItems)
	assert.NotEmpty(t, cacher.setItems["s:f:ns:v-321:flag-1"])
	assert.NotEmpty(t, cacher.setItems["s:n:ns"])
	assert.Equal(t, 2, cacher.setCalled)

	// Second call to get flag should hit the cache
	_, err = cachedStore.GetFlag(context.TODO(), storage.NewResource("ns", "flag-1"))
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.getKeys)
	_, ok := cacher.getKeys["s:f:ns:v-321:flag-1"]
	assert.True(t, ok)

	store.AssertNumberOfCalls(t, "GetFlag", 1)
}

func TestListFlags(t *testing.T) {
	var (
		expectedFlags = storage.ResultSet[*flipt.Flag]{}
		store         = &common.StoreMock{}
	)

	store.On("GetVersion", context.TODO(), storage.NewNamespace("ns")).Return(
		"v-321", nil,
	)
	store.On("ListFlags", context.TODO(),
		mock.Anything,
	).Return(
		expectedFlags, nil,
	)

	var (
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheBackendMemory,
		})
		cacher = newCacheSpy(cache)

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	_, err := cachedStore.ListFlags(context.TODO(), storage.ListWithOptions(
		storage.NewNamespace("ns"),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithLimit(5), storage.WithPageToken("token")),
	))
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.setItems)
	assert.NotEmpty(t, cacher.setItems["s:f:ns:v-321:flags/token-0-5"])
	assert.NotEmpty(t, cacher.setItems["s:n:ns"])
	assert.Equal(t, 2, cacher.setCalled)

	// Second call to get flag should hit the cache
	//
	_, err = cachedStore.ListFlags(context.TODO(), storage.ListWithOptions(
		storage.NewNamespace("ns"),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithLimit(5), storage.WithPageToken("token")),
	))
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.getKeys)
	_, ok := cacher.getKeys["s:f:ns:v-321:flags/token-0-5"]
	assert.True(t, ok)

	store.AssertNumberOfCalls(t, "ListFlags", 1)
}

func TestUpdateNamespace(t *testing.T) {
	var (
		expectedNs = &flipt.Namespace{Key: "ns"}
		store      = &common.StoreMock{}
	)
	store.On("GetVersion", context.TODO(), storage.NewNamespace("ns")).Return(
		"v-321", nil,
	)

	store.On("UpdateNamespace", context.TODO(), mock.Anything).Return(expectedNs, nil)

	var (
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheBackendMemory,
		})
		cacher = newCacheSpy(cache)

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	// Update flag should call the store and delete the cache
	_, err := cachedStore.UpdateNamespace(context.TODO(), &flipt.UpdateNamespaceRequest{Key: "ns"})
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.setItems)
	_, ok := cacher.setItems["s:n:ns"]
	assert.True(t, ok)
	assert.Equal(t, 1, cacher.setCalled)

	store.AssertNumberOfCalls(t, "UpdateNamespace", 1)
}

func TestDeleteNamespace(t *testing.T) {
	store := &common.StoreMock{}

	store.On("DeleteNamespace", context.TODO(), mock.Anything).Return(nil)

	var (
		cache = memory.NewCache(config.CacheConfig{
			TTL:     time.Second,
			Enabled: true,
			Backend: config.CacheBackendMemory,
		})
		cacher = newCacheSpy(cache)

		logger      = zaptest.NewLogger(t)
		cachedStore = NewStore(store, cacher, logger)
	)

	// Delete variant should call the store and delete the cache
	err := cachedStore.DeleteNamespace(context.TODO(), &flipt.DeleteNamespaceRequest{Key: "ns"})
	require.NoError(t, err)
	assert.NotEmpty(t, cacher.deleteKeys)
	_, ok := cacher.deleteKeys["s:n:ns"]
	assert.True(t, ok)
	assert.Equal(t, 1, cacher.deleteCalled)

	store.AssertNumberOfCalls(t, "DeleteNamespace", 1)
}

func TestCUD(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(store *common.StoreMock)
		docall func(store *Store) error
	}{
		{
			"CreateFlag",
			func(store *common.StoreMock) {
				store.On("CreateFlag", context.TODO(), mock.Anything).Return(&flipt.Flag{NamespaceKey: "ns", Key: "flag-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{NamespaceKey: "ns", Key: "flag-1"})
				return err
			},
		},
		{
			"UpdateFlag",
			func(store *common.StoreMock) {
				store.On("UpdateFlag", context.TODO(), mock.Anything).Return(&flipt.Flag{NamespaceKey: "ns", Key: "flag-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{NamespaceKey: "ns", Key: "flag-1"})
				return err
			},
		},
		{
			"DeleteFlag",
			func(store *common.StoreMock) {
				store.On("DeleteFlag", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{NamespaceKey: "ns", Key: "flag-1"})
			},
		},
		{
			"CreateVariant",
			func(store *common.StoreMock) {
				store.On("CreateVariant", context.TODO(), mock.Anything).Return(&flipt.Variant{NamespaceKey: "ns", FlagKey: "flag-1", Key: "var-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{NamespaceKey: "ns", FlagKey: "flag-1"})
				return err
			},
		},
		{
			"UpdateVariant",
			func(store *common.StoreMock) {
				store.On("UpdateVariant", context.TODO(), mock.Anything).Return(&flipt.Variant{NamespaceKey: "ns", FlagKey: "flag-1", Key: "var-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{NamespaceKey: "ns", FlagKey: "flag-1", Id: "var-1"})
				return err
			},
		},
		{
			"DeleteVariant",
			func(store *common.StoreMock) {
				store.On("DeleteVariant", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{NamespaceKey: "ns", FlagKey: "flag-1", Id: "var-1"})
			},
		},
		{
			"CreateSegment",
			func(store *common.StoreMock) {
				store.On("CreateSegment", context.TODO(), mock.Anything).Return(&flipt.Segment{NamespaceKey: "ns", Key: "seg-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{NamespaceKey: "ns", Key: "seg-1"})
				return err
			},
		},
		{
			"UpdateSegment",
			func(store *common.StoreMock) {
				store.On("UpdateSegment", context.TODO(), mock.Anything).Return(&flipt.Segment{NamespaceKey: "ns", Key: "seg-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.UpdateSegment(context.TODO(), &flipt.UpdateSegmentRequest{NamespaceKey: "ns", Key: "seg-1"})
				return err
			},
		},
		{
			"DeleteSegment",
			func(store *common.StoreMock) {
				store.On("DeleteSegment", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.DeleteSegment(context.TODO(), &flipt.DeleteSegmentRequest{NamespaceKey: "ns", Key: "seg-1"})
			},
		},
		{
			"CreateConstraint",
			func(store *common.StoreMock) {
				store.On("CreateConstraint", context.TODO(), mock.Anything).Return(&flipt.Constraint{NamespaceKey: "ns", SegmentKey: "seg-1", Id: "a1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{NamespaceKey: "ns", SegmentKey: "seg-1"})
				return err
			},
		},
		{
			"UpdateConstraint",
			func(store *common.StoreMock) {
				store.On("UpdateConstraint", context.TODO(), mock.Anything).Return(&flipt.Constraint{NamespaceKey: "ns", SegmentKey: "seg-1", Id: "a1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.UpdateConstraint(context.TODO(), &flipt.UpdateConstraintRequest{NamespaceKey: "ns", SegmentKey: "seg-1", Id: "a1"})
				return err
			},
		},
		{
			"DeleteConstraint",
			func(store *common.StoreMock) {
				store.On("DeleteConstraint", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.DeleteConstraint(context.TODO(), &flipt.DeleteConstraintRequest{NamespaceKey: "ns", SegmentKey: "seg-1", Id: "a1"})
			},
		},
		{
			"CreateRule",
			func(store *common.StoreMock) {
				store.On("CreateRule", context.TODO(), mock.Anything).Return(&flipt.Rule{NamespaceKey: "ns", FlagKey: "flag-1", Id: "rule-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.CreateRule(context.TODO(), &flipt.CreateRuleRequest{NamespaceKey: "ns", FlagKey: "flag-1"})
				return err
			},
		},
		{
			"UpdateRule",
			func(store *common.StoreMock) {
				store.On("UpdateRule", context.TODO(), mock.Anything).Return(&flipt.Rule{NamespaceKey: "ns", FlagKey: "flag-1", Id: "rule-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{NamespaceKey: "ns", FlagKey: "rule-1", Id: "rule-1"})
				return err
			},
		},
		{
			"DeleteRule",
			func(store *common.StoreMock) {
				store.On("DeleteRule", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{NamespaceKey: "ns", FlagKey: "flag-1", Id: "rule-1"})
			},
		},
		{
			"OrderRules",
			func(store *common.StoreMock) {
				store.On("OrderRules", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.OrderRules(context.TODO(), &flipt.OrderRulesRequest{NamespaceKey: "ns", FlagKey: "flag-1", RuleIds: []string{"rule-1"}})
			},
		},
		{
			"CreateDistribution",
			func(store *common.StoreMock) {
				store.On("CreateDistribution", context.TODO(), mock.Anything).Return(&flipt.Distribution{RuleId: "rule-1", Id: "dist-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{NamespaceKey: "ns", RuleId: "rule-1"})
				return err
			},
		},
		{
			"UpdateDistribution",
			func(store *common.StoreMock) {
				store.On("UpdateDistribution", context.TODO(), mock.Anything).Return(&flipt.Distribution{RuleId: "rule-1", Id: "dist-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{NamespaceKey: "ns", RuleId: "rule-1", Id: "dist-1"})
				return err
			},
		},
		{
			"DeleteDistribution",
			func(store *common.StoreMock) {
				store.On("DeleteDistribution", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.DeleteDistribution(context.TODO(), &flipt.DeleteDistributionRequest{NamespaceKey: "ns", FlagKey: "flag-1", Id: "dist-1"})
			},
		},
		{
			"CreateRollout",
			func(store *common.StoreMock) {
				store.On("CreateRollout", context.TODO(), mock.Anything).Return(&flipt.Rollout{NamespaceKey: "ns", FlagKey: "flag-1", Id: "roll-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{NamespaceKey: "ns", FlagKey: "flag-1"})
				return err
			},
		},
		{
			"UpdateRollout",
			func(store *common.StoreMock) {
				store.On("UpdateRollout", context.TODO(), mock.Anything).Return(&flipt.Rollout{NamespaceKey: "ns", FlagKey: "flag-1", Id: "roll-1"}, nil)
			},
			func(cachedStore *Store) error {
				_, err := cachedStore.UpdateRollout(context.TODO(), &flipt.UpdateRolloutRequest{NamespaceKey: "ns", FlagKey: "flag-1", Id: "roll-1"})
				return err
			},
		},
		{
			"DeleteRollout",
			func(store *common.StoreMock) {
				store.On("DeleteRollout", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.DeleteRollout(context.TODO(), &flipt.DeleteRolloutRequest{NamespaceKey: "ns", FlagKey: "flag-1", Id: "roll-1"})
			},
		},
		{
			"OrderRollouts",
			func(store *common.StoreMock) {
				store.On("OrderRollouts", context.TODO(), mock.Anything).Return(nil)
			},
			func(cachedStore *Store) error {
				return cachedStore.OrderRollouts(context.TODO(), &flipt.OrderRolloutsRequest{NamespaceKey: "ns", FlagKey: "flag-1", RolloutIds: []string{"roll-1"}})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &common.StoreMock{}

			store.On("GetVersion", context.TODO(), storage.NewNamespace("ns")).Return(
				"v-321", nil,
			)
			tt.setup(store)

			var (
				cache = memory.NewCache(config.CacheConfig{
					TTL:     time.Second,
					Enabled: true,
					Backend: config.CacheBackendMemory,
				})
				cacher = newCacheSpy(cache)

				logger      = zaptest.NewLogger(t)
				cachedStore = NewStore(store, cacher, logger)
			)

			err := tt.docall(cachedStore)
			require.NoError(t, err)
			assert.NotEmpty(t, cacher.setItems)
			_, ok := cacher.setItems["s:n:ns"]
			assert.True(t, ok)
			assert.Equal(t, 1, cacher.setCalled)

			store.AssertNumberOfCalls(t, tt.name, 1)
		})
	}
}
