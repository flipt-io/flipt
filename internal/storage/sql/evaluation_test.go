package sql_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/server"
	"go.flipt.io/flipt/internal/server/evaluation"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
)

func (s *DBTestSuite) TestGetEvaluationRules() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	// constraint 1
	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)

	// constraint 2
	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foz",
		Operator:   "EQ",
		Value:      "baz",
	})

	require.NoError(t, err)

	// rule rank 1
	rule1, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:     flag.Key,
		SegmentKeys: []string{segment.Key},
		Rank:        1,
	})

	require.NoError(t, err)

	// rule rank 2
	rule2, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:     flag.Key,
		SegmentKeys: []string{segment.Key},
		Rank:        2,
	})

	require.NoError(t, err)

	evaluationRules, err := s.store.GetEvaluationRules(context.TODO(), storage.NewResource(storage.DefaultNamespace, flag.Key))
	require.NoError(t, err)

	assert.NotEmpty(t, evaluationRules)
	assert.Len(t, evaluationRules, 2)

	assert.Equal(t, rule1.Id, evaluationRules[0].ID)
	assert.Equal(t, storage.DefaultNamespace, evaluationRules[0].NamespaceKey)
	assert.Equal(t, rule1.FlagKey, evaluationRules[0].FlagKey)

	assert.Equal(t, rule1.SegmentKey, evaluationRules[0].Segments[segment.Key].SegmentKey)
	assert.Equal(t, segment.MatchType, evaluationRules[0].Segments[segment.Key].MatchType)
	assert.Equal(t, rule1.Rank, evaluationRules[0].Rank)
	assert.Len(t, evaluationRules[0].Segments[segment.Key].Constraints, 2)

	assert.Equal(t, rule2.Id, evaluationRules[1].ID)
	assert.Equal(t, storage.DefaultNamespace, evaluationRules[1].NamespaceKey)
	assert.Equal(t, rule2.FlagKey, evaluationRules[1].FlagKey)

	assert.Equal(t, rule2.SegmentKey, evaluationRules[1].Segments[segment.Key].SegmentKey)
	assert.Equal(t, segment.MatchType, evaluationRules[1].Segments[segment.Key].MatchType)
	assert.Equal(t, rule2.Rank, evaluationRules[1].Rank)
	assert.Len(t, evaluationRules[1].Segments[segment.Key].Constraints, 2)
}

func (s *DBTestSuite) TestGetEvaluationRules_NoNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	// constraint 1
	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)

	// constraint 2
	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foz",
		Operator:   "EQ",
		Value:      "baz",
	})

	require.NoError(t, err)

	// rule rank 1
	rule1, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:     flag.Key,
		SegmentKeys: []string{segment.Key},
		Rank:        1,
	})

	require.NoError(t, err)

	// rule rank 2
	rule2, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:     flag.Key,
		SegmentKeys: []string{segment.Key},
		Rank:        2,
	})

	require.NoError(t, err)

	evaluationRules, err := s.store.GetEvaluationRules(context.TODO(), storage.NewResource("", flag.Key))
	require.NoError(t, err)

	assert.NotEmpty(t, evaluationRules)
	assert.Len(t, evaluationRules, 2)

	assert.Equal(t, rule1.Id, evaluationRules[0].ID)
	assert.Equal(t, storage.DefaultNamespace, evaluationRules[0].NamespaceKey)
	assert.Equal(t, rule1.FlagKey, evaluationRules[0].FlagKey)

	assert.Equal(t, rule1.SegmentKey, evaluationRules[0].Segments[segment.Key].SegmentKey)
	assert.Equal(t, segment.MatchType, evaluationRules[0].Segments[segment.Key].MatchType)
	assert.Equal(t, rule1.Rank, evaluationRules[0].Rank)
	assert.Len(t, evaluationRules[0].Segments[segment.Key].Constraints, 2)

	assert.Equal(t, rule2.Id, evaluationRules[1].ID)
	assert.Equal(t, storage.DefaultNamespace, evaluationRules[1].NamespaceKey)
	assert.Equal(t, rule2.FlagKey, evaluationRules[1].FlagKey)

	assert.Equal(t, rule2.SegmentKey, evaluationRules[1].Segments[segment.Key].SegmentKey)
	assert.Equal(t, segment.MatchType, evaluationRules[1].Segments[segment.Key].MatchType)
	assert.Equal(t, rule2.Rank, evaluationRules[1].Rank)
	assert.Len(t, evaluationRules[1].Segments[segment.Key].Constraints, 2)
}

func (s *DBTestSuite) TestGetEvaluationRulesNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)

	firstSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	secondSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          "another_segment",
		Name:         "foo",
		Description:  "bar",
		MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	// constraint 1
	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   firstSegment.Key,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "EQ",
		Value:        "bar",
	})

	require.NoError(t, err)

	// constraint 2
	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   firstSegment.Key,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foz",
		Operator:     "EQ",
		Value:        "baz",
	})

	require.NoError(t, err)

	// rule rank 1
	rule1, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		NamespaceKey:    s.namespace,
		FlagKey:         flag.Key,
		SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
		SegmentKeys:     []string{firstSegment.Key, secondSegment.Key},
		Rank:            1,
	})

	require.NoError(t, err)

	evaluationRules, err := s.store.GetEvaluationRules(context.TODO(), storage.NewResource(s.namespace, flag.Key))
	require.NoError(t, err)

	assert.NotEmpty(t, evaluationRules)
	assert.Len(t, evaluationRules, 1)

	assert.Equal(t, rule1.Id, evaluationRules[0].ID)
	assert.Equal(t, s.namespace, evaluationRules[0].NamespaceKey)
	assert.Equal(t, rule1.FlagKey, evaluationRules[0].FlagKey)

	assert.Equal(t, firstSegment.Key, evaluationRules[0].Segments[firstSegment.Key].SegmentKey)
	assert.Equal(t, firstSegment.MatchType, evaluationRules[0].Segments[firstSegment.Key].MatchType)
	assert.Equal(t, rule1.Rank, evaluationRules[0].Rank)
	assert.Len(t, evaluationRules[0].Segments[firstSegment.Key].Constraints, 2)

	assert.Equal(t, secondSegment.Key, evaluationRules[0].Segments[secondSegment.Key].SegmentKey)
	assert.Equal(t, secondSegment.MatchType, evaluationRules[0].Segments[secondSegment.Key].MatchType)
	assert.Equal(t, rule1.Rank, evaluationRules[0].Rank)
	assert.Empty(t, evaluationRules[0].Segments[secondSegment.Key].Constraints)
}

func (s *DBTestSuite) TestGetEvaluationDistributions() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	// variant 1
	variant1, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey: flag.Key,
		Key:     "foo",
	})

	require.NoError(t, err)

	// variant 2
	variant2, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey:    flag.Key,
		Key:        "bar",
		Attachment: `{"key2":   "value2"}`,
	})

	require.NoError(t, err)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)

	// 50/50 distribution
	_, err = s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant1.Id,
		Rollout:   50.00,
	})

	// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
	time.Sleep(1 * time.Second)

	require.NoError(t, err)

	_, err = s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant2.Id,
		Rollout:   50.00,
	})

	require.NoError(t, err)

	evaluationDistributions, err := s.store.GetEvaluationDistributions(context.TODO(), storage.NewID(rule.Id))
	require.NoError(t, err)

	assert.Len(t, evaluationDistributions, 2)

	assert.NotEmpty(t, evaluationDistributions[0].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[0].RuleID)
	assert.Equal(t, variant1.Id, evaluationDistributions[0].VariantID)
	assert.Equal(t, variant1.Key, evaluationDistributions[0].VariantKey)
	assert.InDelta(t, 50.00, evaluationDistributions[0].Rollout, 0)

	assert.NotEmpty(t, evaluationDistributions[1].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[1].RuleID)
	assert.Equal(t, variant2.Id, evaluationDistributions[1].VariantID)
	assert.Equal(t, variant2.Key, evaluationDistributions[1].VariantKey)
	assert.Equal(t, `{"key2":"value2"}`, evaluationDistributions[1].VariantAttachment)
	assert.InDelta(t, 50.00, evaluationDistributions[1].Rollout, 0)
}

func (s *DBTestSuite) TestGetEvaluationDistributionsNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
	})

	require.NoError(t, err)

	// variant 1
	variant1, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          "foo",
	})

	require.NoError(t, err)

	// variant 2
	variant2, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Key:          "bar",
		Attachment:   `{"key2":   "value2"}`,
	})

	require.NoError(t, err)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   segment.Key,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "EQ",
		Value:        "bar",
	})

	require.NoError(t, err)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		SegmentKey:   segment.Key,
		Rank:         1,
	})

	require.NoError(t, err)

	// 50/50 distribution
	_, err = s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		RuleId:       rule.Id,
		VariantId:    variant1.Id,
		Rollout:      50.00,
	})

	// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
	time.Sleep(1 * time.Second)

	require.NoError(t, err)

	_, err = s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		RuleId:       rule.Id,
		VariantId:    variant2.Id,
		Rollout:      50.00,
	})

	require.NoError(t, err)

	evaluationDistributions, err := s.store.GetEvaluationDistributions(context.TODO(), storage.NewID(rule.Id))
	require.NoError(t, err)

	assert.Len(t, evaluationDistributions, 2)

	assert.NotEmpty(t, evaluationDistributions[0].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[0].RuleID)
	assert.Equal(t, variant1.Id, evaluationDistributions[0].VariantID)
	assert.Equal(t, variant1.Key, evaluationDistributions[0].VariantKey)
	assert.InDelta(t, 50.00, evaluationDistributions[0].Rollout, 0)

	assert.NotEmpty(t, evaluationDistributions[1].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[1].RuleID)
	assert.Equal(t, variant2.Id, evaluationDistributions[1].VariantID)
	assert.Equal(t, variant2.Key, evaluationDistributions[1].VariantKey)
	assert.Equal(t, `{"key2":"value2"}`, evaluationDistributions[1].VariantAttachment)
	assert.InDelta(t, 50.00, evaluationDistributions[1].Rollout, 0)
}

// https://github.com/flipt-io/flipt/issues/229
func (s *DBTestSuite) TestGetEvaluationDistributions_MaintainOrder() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
	})

	require.NoError(t, err)

	// variant 1
	variant1, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey: flag.Key,
		Key:     "foo",
	})

	require.NoError(t, err)

	// variant 2
	variant2, err := s.store.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{
		FlagKey: flag.Key,
		Key:     "bar",
	})

	require.NoError(t, err)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: segment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)

	rule, err := s.store.CreateRule(context.TODO(), &flipt.CreateRuleRequest{
		FlagKey:    flag.Key,
		SegmentKey: segment.Key,
		Rank:       1,
	})

	require.NoError(t, err)

	// 80/20 distribution
	dist1, err := s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant1.Id,
		Rollout:   80.00,
	})

	require.NoError(t, err)

	// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
	time.Sleep(1 * time.Second)

	dist2, err := s.store.CreateDistribution(context.TODO(), &flipt.CreateDistributionRequest{
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant2.Id,
		Rollout:   20.00,
	})

	require.NoError(t, err)

	evaluationDistributions, err := s.store.GetEvaluationDistributions(context.TODO(), storage.NewID(rule.Id))
	require.NoError(t, err)

	assert.Len(t, evaluationDistributions, 2)

	assert.NotEmpty(t, evaluationDistributions[0].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[0].RuleID)
	assert.Equal(t, variant1.Id, evaluationDistributions[0].VariantID)
	assert.Equal(t, variant1.Key, evaluationDistributions[0].VariantKey)
	assert.InDelta(t, 80.00, evaluationDistributions[0].Rollout, 0)

	assert.NotEmpty(t, evaluationDistributions[1].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[1].RuleID)
	assert.Equal(t, variant2.Id, evaluationDistributions[1].VariantID)
	assert.Equal(t, variant2.Key, evaluationDistributions[1].VariantKey)
	assert.InDelta(t, 20.00, evaluationDistributions[1].Rollout, 0)

	// update dist1 with same values
	_, err = s.store.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
		Id:        dist1.Id,
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant1.Id,
		Rollout:   80.00,
	})

	require.NoError(t, err)

	// required for MySQL since it only s.stores timestamps to the second and not millisecond granularity
	time.Sleep(1 * time.Second)

	// update dist2 with same values
	_, err = s.store.UpdateDistribution(context.TODO(), &flipt.UpdateDistributionRequest{
		Id:        dist2.Id,
		FlagKey:   flag.Key,
		RuleId:    rule.Id,
		VariantId: variant2.Id,
		Rollout:   20.00,
	})

	require.NoError(t, err)

	evaluationDistributions, err = s.store.GetEvaluationDistributions(context.TODO(), storage.NewID(rule.Id))
	require.NoError(t, err)

	assert.Len(t, evaluationDistributions, 2)

	assert.NotEmpty(t, evaluationDistributions[0].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[0].RuleID)
	assert.Equal(t, variant1.Id, evaluationDistributions[0].VariantID)
	assert.Equal(t, variant1.Key, evaluationDistributions[0].VariantKey)
	assert.InDelta(t, 80.00, evaluationDistributions[0].Rollout, 0)

	assert.NotEmpty(t, evaluationDistributions[1].ID)
	assert.Equal(t, rule.Id, evaluationDistributions[1].RuleID)
	assert.Equal(t, variant2.Id, evaluationDistributions[1].VariantID)
	assert.Equal(t, variant2.Key, evaluationDistributions[1].VariantKey)
	assert.InDelta(t, 20.00, evaluationDistributions[1].Rollout, 0)
}

func (s *DBTestSuite) TestGetEvaluationRollouts() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	_, err = s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: flag.Key,
		Rank:    1,
		Rule: &flipt.CreateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Percentage: 50.0,
				Value:      false,
			},
		},
	})

	require.NoError(t, err)

	_, err = s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: flag.Key,
		Rank:    2,
		Rule: &flipt.CreateRolloutRequest_Segment{
			Segment: &flipt.RolloutSegment{
				SegmentKeys: []string{segment.Key},
				Value:       true,
			},
		},
	})

	require.NoError(t, err)

	evaluationRollouts, err := s.store.GetEvaluationRollouts(context.TODO(), storage.NewResource(storage.DefaultNamespace, flag.Key))
	require.NoError(t, err)

	assert.Len(t, evaluationRollouts, 2)

	assert.Equal(t, "default", evaluationRollouts[0].NamespaceKey)
	assert.Equal(t, int32(1), evaluationRollouts[0].Rank)
	assert.NotNil(t, evaluationRollouts[0].Threshold)
	assert.InDelta(t, 50.0, evaluationRollouts[0].Threshold.Percentage, 0)
	assert.False(t, evaluationRollouts[0].Threshold.Value, "percentage value is false")

	assert.Equal(t, "default", evaluationRollouts[1].NamespaceKey)
	assert.Equal(t, int32(2), evaluationRollouts[1].Rank)
	assert.NotNil(t, evaluationRollouts[1].Segment)

	assert.Contains(t, evaluationRollouts[1].Segment.Segments, segment.Key)
	assert.Equal(t, segment.MatchType, evaluationRollouts[1].Segment.Segments[segment.Key].MatchType)

	assert.True(t, evaluationRollouts[1].Segment.Value, "segment value is true")
}

func (s *DBTestSuite) TestGetEvaluationRollouts_NoNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		Enabled:     true,
		Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)

	firstSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         t.Name(),
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	secondSegment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		Key:         "another_segment",
		Name:        "foo",
		Description: "bar",
		MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		SegmentKey: firstSegment.Key,
		Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:   "foo",
		Operator:   "EQ",
		Value:      "bar",
	})

	require.NoError(t, err)

	_, err = s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: flag.Key,
		Rank:    1,
		Rule: &flipt.CreateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Percentage: 50.0,
				Value:      false,
			},
		},
	})

	require.NoError(t, err)

	_, err = s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		FlagKey: flag.Key,
		Rank:    2,
		Rule: &flipt.CreateRolloutRequest_Segment{
			Segment: &flipt.RolloutSegment{
				SegmentKeys:     []string{firstSegment.Key, secondSegment.Key},
				SegmentOperator: flipt.SegmentOperator_AND_SEGMENT_OPERATOR,
				Value:           true,
			},
		},
	})

	require.NoError(t, err)

	evaluationRollouts, err := s.store.GetEvaluationRollouts(context.TODO(), storage.NewResource("", flag.Key))
	require.NoError(t, err)

	assert.Len(t, evaluationRollouts, 2)

	assert.Equal(t, "default", evaluationRollouts[0].NamespaceKey)
	assert.Equal(t, int32(1), evaluationRollouts[0].Rank)
	assert.NotNil(t, evaluationRollouts[0].Threshold)
	assert.InDelta(t, 50.0, evaluationRollouts[0].Threshold.Percentage, 0)
	assert.False(t, evaluationRollouts[0].Threshold.Value, "percentage value is false")

	assert.Equal(t, "default", evaluationRollouts[1].NamespaceKey)
	assert.Equal(t, int32(2), evaluationRollouts[1].Rank)
	assert.NotNil(t, evaluationRollouts[1].Segment)

	assert.Equal(t, firstSegment.Key, evaluationRollouts[1].Segment.Segments[firstSegment.Key].SegmentKey)
	assert.Equal(t, flipt.SegmentOperator_AND_SEGMENT_OPERATOR, evaluationRollouts[1].Segment.SegmentOperator)
	assert.Len(t, evaluationRollouts[1].Segment.Segments[firstSegment.Key].Constraints, 1)

	assert.Equal(t, secondSegment.Key, evaluationRollouts[1].Segment.Segments[secondSegment.Key].SegmentKey)
	assert.Empty(t, evaluationRollouts[1].Segment.Segments[secondSegment.Key].Constraints)
}

func (s *DBTestSuite) TestGetEvaluationRollouts_NonDefaultNamespace() {
	t := s.T()

	flag, err := s.store.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		Enabled:      true,
		Type:         flipt.FlagType_BOOLEAN_FLAG_TYPE,
	})

	require.NoError(t, err)

	segment, err := s.store.CreateSegment(context.TODO(), &flipt.CreateSegmentRequest{
		NamespaceKey: s.namespace,
		Key:          t.Name(),
		Name:         "foo",
		Description:  "bar",
		MatchType:    flipt.MatchType_ANY_MATCH_TYPE,
	})

	require.NoError(t, err)

	_, err = s.store.CreateConstraint(context.TODO(), &flipt.CreateConstraintRequest{
		NamespaceKey: s.namespace,
		SegmentKey:   segment.Key,
		Type:         flipt.ComparisonType_STRING_COMPARISON_TYPE,
		Property:     "foo",
		Operator:     "EQ",
		Value:        "bar",
	})

	require.NoError(t, err)

	_, err = s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Rank:         1,
		Rule: &flipt.CreateRolloutRequest_Threshold{
			Threshold: &flipt.RolloutThreshold{
				Percentage: 50.0,
				Value:      false,
			},
		},
	})

	require.NoError(t, err)

	_, err = s.store.CreateRollout(context.TODO(), &flipt.CreateRolloutRequest{
		NamespaceKey: s.namespace,
		FlagKey:      flag.Key,
		Rank:         2,
		Rule: &flipt.CreateRolloutRequest_Segment{
			Segment: &flipt.RolloutSegment{
				SegmentKey: segment.Key,
				Value:      true,
			},
		},
	})

	require.NoError(t, err)

	evaluationRollouts, err := s.store.GetEvaluationRollouts(context.TODO(), storage.NewResource(s.namespace, flag.Key))
	require.NoError(t, err)

	assert.Len(t, evaluationRollouts, 2)

	assert.Equal(t, s.namespace, evaluationRollouts[0].NamespaceKey)
	assert.Equal(t, int32(1), evaluationRollouts[0].Rank)
	assert.NotNil(t, evaluationRollouts[0].Threshold)
	assert.InDelta(t, 50.0, evaluationRollouts[0].Threshold.Percentage, 0)
	assert.False(t, evaluationRollouts[0].Threshold.Value, "percentage value is false")

	assert.Equal(t, s.namespace, evaluationRollouts[1].NamespaceKey)
	assert.Equal(t, int32(2), evaluationRollouts[1].Rank)
	assert.NotNil(t, evaluationRollouts[1].Segment)

	assert.Contains(t, evaluationRollouts[1].Segment.Segments, segment.Key)
	assert.Equal(t, segment.MatchType, evaluationRollouts[1].Segment.Segments[segment.Key].MatchType)

	assert.True(t, evaluationRollouts[1].Segment.Value, "segment value is true")
}

func Benchmark_EvaluationV1AndV2(b *testing.B) {
	s := new(DBTestSuite)

	t := &testing.T{}
	s.SetT(t)
	s.SetupSuite()

	server := server.New(zap.NewNop(), s.store)
	importer := ext.NewImporter(server)
	reader, err := os.Open("./testdata/benchmark_test.yml")
	defer func() {
		_ = reader.Close()
	}()
	skipExistingFalse := false

	require.NoError(t, err)

	err = importer.Import(context.TODO(), ext.EncodingYML, reader, skipExistingFalse)
	require.NoError(t, err)

	flagKeys := make([]string, 0, 10)

	for i := 0; i < 10; i++ {
		var flagKey string

		num := rand.Intn(50) + 1

		if num < 10 {
			flagKey = fmt.Sprintf("flag_00%d", num)
		} else {
			flagKey = fmt.Sprintf("flag_0%d", num)
		}

		flagKeys = append(flagKeys, flagKey)
	}

	eserver := evaluation.New(zap.NewNop(), s.store)

	b.ResetTimer()

	for _, flagKey := range flagKeys {
		entityId := uuid.Must(uuid.NewV4()).String()
		ereq := &flipt.EvaluationRequest{
			FlagKey:  flagKey,
			EntityId: entityId,
		}

		ev2req := &rpcevaluation.EvaluationRequest{
			FlagKey:  flagKey,
			EntityId: entityId,
		}

		b.Run(fmt.Sprintf("evaluation-v1-%s", flagKey), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				evaluation, err := server.Evaluate(context.TODO(), ereq)

				require.NoError(t, err)
				assert.NotEmpty(t, evaluation)
			}
		})

		b.Run(fmt.Sprintf("variant-evaluation-%s", flagKey), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				variant, err := eserver.Variant(context.TODO(), ev2req)

				require.NoError(t, err)
				assert.NotEmpty(t, variant)
			}
		})
	}

	for _, flagKey := range []string{"flag_boolean", "another_boolean_flag"} {
		breq := &rpcevaluation.EvaluationRequest{
			FlagKey:  flagKey,
			EntityId: uuid.Must(uuid.NewV4()).String(),
		}

		b.Run(fmt.Sprintf("boolean-evaluation-%s", flagKey), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				boolean, err := eserver.Boolean(context.TODO(), breq)

				require.NoError(t, err)
				assert.NotEmpty(t, boolean)
			}
		})
	}

	s.TearDownSuite()
}
