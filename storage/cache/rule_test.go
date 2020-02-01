package cache

import (
	"context"
	"testing"

	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetRule(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	store.On("GetRule", mock.Anything, mock.Anything).Return(&flipt.Rule{Id: "foo"}, nil)
	cacher.On("Get", mock.Anything).Return(&flipt.Rule{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.GetRule(context.TODO(), &flipt.GetRuleRequest{Id: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", "rule:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", "rule:foo")

	cacher.On("Get", mock.Anything).Return(&flipt.Rule{Id: "foo"}, true)

	got, err = subject.GetRule(context.TODO(), &flipt.GetRuleRequest{Id: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetRuleNotFound(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	store.On("GetRule", mock.Anything, mock.Anything).Return(&flipt.Rule{}, errors.ErrNotFound("foo"))
	cacher.On("Get", mock.Anything).Return(&flipt.Rule{}, false).Once()

	_, err := subject.GetRule(context.TODO(), &flipt.GetRuleRequest{Id: "foo"})
	require.Error(t, err)

	// doesnt exists so it should not be added
	cacher.AssertNotCalled(t, "Set")
	cacher.AssertCalled(t, "Get", "rule:foo")
}

func TestListRules(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	ret := []*flipt.Rule{
		{Id: "foo"},
		{Id: "bar"},
	}

	store.On("ListRules", mock.Anything, mock.Anything).Return(ret, nil)
	cacher.On("Get", mock.Anything).Return([]*flipt.Rule{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.ListRules(context.TODO(), &flipt.ListRuleRequest{FlagKey: "foo"})
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", "rules:flag:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", "rules:flag:foo")

	cacher.On("Get", mock.Anything).Return(ret, true)

	got, err = subject.ListRules(context.TODO(), &flipt.ListRuleRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestListRules_NoResults(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	ret := []*flipt.Rule{}

	store.On("ListRules", mock.Anything, mock.Anything).Return(ret, nil)
	cacher.On("Get", mock.Anything).Return([]*flipt.Rule{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.ListRules(context.TODO(), &flipt.ListRuleRequest{FlagKey: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt be set in the cahe
	cacher.AssertNotCalled(t, "Set")
	cacher.AssertCalled(t, "Get", "rules:flag:foo")

	cacher.On("Get", mock.Anything).Return(ret, true)

	got, err = subject.ListRules(context.TODO(), &flipt.ListRuleRequest{})
	require.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateRule(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	store.On("CreateRule", mock.Anything, mock.Anything).Return(&flipt.Rule{Id: "foo"}, nil)
	cacher.On("Flush")

	rule, err := subject.CreateRule(context.TODO(), &flipt.CreateRuleRequest{FlagKey: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, rule)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestUpdateRule(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	store.On("UpdateRule", mock.Anything, mock.Anything).Return(&flipt.Rule{Id: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.UpdateRule(context.TODO(), &flipt.UpdateRuleRequest{Id: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestDeleteRule(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	store.On("DeleteRule", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush")

	err := subject.DeleteRule(context.TODO(), &flipt.DeleteRuleRequest{Id: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestCreateDistribution(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	store.On("CreateDistribution", mock.Anything, mock.Anything).Return(&flipt.Distribution{RuleId: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{RuleId: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestUpdateDistribution(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	store.On("UpdateDistribution", mock.Anything, mock.Anything).Return(&flipt.Distribution{RuleId: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{RuleId: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestDeleteDistribution(t *testing.T) {
	var (
		store   = &ruleStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewRuleCache(logger, cacher, store)
	)

	store.On("DeleteDistribution", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush")

	err := subject.DeleteDistribution(context.TODO(), &flipt.DeleteDistributionRequest{RuleId: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}
