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

	ret := []*storage.EvaluationRule{}

	store.On("GetEvaluationRules", mock.Anything, mock.Anything).Return(ret, nil)
	cacher.On("Get", mock.Anything).Return([]*storage.EvaluationRule{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.GetEvaluationRules(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", "e:r:f:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", "e:r:f:foo")

	cacher.On("Get", mock.Anything).Return(ret, true)

	got, err = subject.GetEvaluationRules(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetEvaluationDistributions(t *testing.T) {
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

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", "e:d:r:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", "e:d:r:foo")

	cacher.On("Get", mock.Anything).Return(ret, true)

	got, err = subject.GetEvaluationDistributions(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}
