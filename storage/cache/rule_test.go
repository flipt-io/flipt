package cache

import (
	"context"
	"testing"

	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetRule(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("GetRule", mock.Anything, mock.Anything).Return(&flipt.Rule{Id: "foo"}, nil)
	cacher.On("Get", mock.Anything, mock.Anything).Return(&flipt.Rule{}, false, nil).Once()
	cacher.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	got, err := subject.GetRule(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", mock.Anything, "rule:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", mock.Anything, "rule:foo")

	cacher.On("Get", mock.Anything, mock.Anything).Return(&flipt.Rule{Id: "foo"}, true, nil)

	got, err = subject.GetRule(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetRuleNotFound(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("GetRule", mock.Anything, mock.Anything).Return(&flipt.Rule{}, errors.ErrNotFound("foo"))
	cacher.On("Get", mock.Anything, mock.Anything).Return(&flipt.Rule{}, false, nil)

	_, err := subject.GetRule(context.TODO(), "foo")
	require.Error(t, err)

	// doesnt exists so it should not be added
	cacher.AssertNotCalled(t, "Set")
	cacher.AssertCalled(t, "Get", mock.Anything, "rule:foo")
}

func TestListRules(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	ret := []*flipt.Rule{
		{Id: "foo"},
		{Id: "bar"},
	}

	store.On("ListRules", mock.Anything, "foo", mock.Anything).Return(ret, nil)

	got, err := subject.ListRules(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	cacher.AssertNotCalled(t, "Get")
	cacher.AssertNotCalled(t, "Set")
}

func TestCreateRule(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("CreateRule", mock.Anything, mock.Anything).Return(&flipt.Rule{Id: "foo"}, nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	rule, err := subject.CreateRule(context.TODO(), &flipt.CreateRuleRequest{FlagKey: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, rule)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestUpdateRule(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("UpdateRule", mock.Anything, mock.Anything).Return(&flipt.Rule{Id: "foo"}, nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	_, err := subject.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{Id: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestDeleteRule(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("DeleteRule", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	err := subject.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{Id: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestCreateDistribution(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("CreateDistribution", mock.Anything, mock.Anything).Return(&flipt.Distribution{RuleId: "foo"}, nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	_, err := subject.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{RuleId: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestUpdateDistribution(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("UpdateDistribution", mock.Anything, mock.Anything).Return(&flipt.Distribution{RuleId: "foo"}, nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	_, err := subject.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{RuleId: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestDeleteDistribution(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("DeleteDistribution", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	err := subject.DeleteDistribution(context.TODO(), &flipt.DeleteDistributionRequest{RuleId: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}
