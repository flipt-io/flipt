package cache

import (
	"context"
	"testing"

	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetEvaluationRules(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &evaluationStoreMock{}
		cacher    = &cacherSpy{}
		subject   = NewEvaluationCache(logger, cacher, store)
	)

	ret := []*storage.EvaluationRule{
		{
			ID: "123",
		},
		{
			ID: "456",
		},
	}

	store.On("GetEvaluationRules", mock.Anything, mock.Anything).Return(ret, nil)
	cacher.On("Get", mock.Anything).Return([]*storage.EvaluationRule{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.GetEvaluationRules(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", "eval:rules:flag:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", "eval:rules:flag:foo")

	cacher.On("Get", mock.Anything).Return(ret, true)

	got, err = subject.GetEvaluationRules(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetEvaluationRules_NoResults(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &evaluationStoreMock{}
		cacher    = &cacherSpy{}
		subject   = NewEvaluationCache(logger, cacher, store)
	)

	ret := []*storage.EvaluationRule{}

	store.On("GetEvaluationRules", mock.Anything, mock.Anything).Return(ret, nil)
	cacher.On("Get", mock.Anything).Return([]*storage.EvaluationRule{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.GetEvaluationRules(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should not be set in the cache
	cacher.AssertNotCalled(t, "Set")
	cacher.AssertCalled(t, "Get", "eval:rules:flag:foo")

	cacher.On("Get", mock.Anything).Return(ret, true)

	got, err = subject.GetEvaluationRules(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestGetEvaluationDistributions(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &evaluationStoreMock{}
		cacher    = &cacherSpy{}
		subject   = NewEvaluationCache(logger, cacher, store)
	)

	ret := []*storage.EvaluationDistribution{
		{
			ID: "123",
		},
		{
			ID: "456",
		},
	}

	store.On("GetEvaluationDistributions", mock.Anything, mock.Anything).Return(ret, nil)
	cacher.On("Get", mock.Anything).Return([]*storage.EvaluationDistribution{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.GetEvaluationDistributions(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", "eval:dist:rule:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", "eval:dist:rule:foo")

	cacher.On("Get", mock.Anything).Return(ret, true)

	got, err = subject.GetEvaluationDistributions(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetEvaluationDistributions_NoResults(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &evaluationStoreMock{}
		cacher    = &cacherSpy{}
		subject   = NewEvaluationCache(logger, cacher, store)
	)

	ret := []*storage.EvaluationDistribution{}

	store.On("GetEvaluationDistributions", mock.Anything, mock.Anything).Return(ret, nil)
	cacher.On("Get", mock.Anything).Return([]*storage.EvaluationDistribution{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.GetEvaluationDistributions(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should not be set in the cache
	cacher.AssertNotCalled(t, "Set")
	cacher.AssertCalled(t, "Get", "eval:dist:rule:foo")

	cacher.On("Get", mock.Anything).Return(ret, true)

	got, err = subject.GetEvaluationDistributions(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)
}
