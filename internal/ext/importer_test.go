package ext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
)

var (
	extensions        = []Encoding{EncodingYML, EncodingJSON}
	skipExistingFalse = false
)

type mockCreator struct {
	getNSReqs []*flipt.GetNamespaceRequest
	getNSErr  error

	createNSReqs []*flipt.CreateNamespaceRequest
	createNSErr  error

	createflagReqs []*flipt.CreateFlagRequest
	createflagErr  error

	updateFlagReqs []*flipt.UpdateFlagRequest
	updateFlagErr  error

	variantReqs []*flipt.CreateVariantRequest
	variantErr  error

	segmentReqs []*flipt.CreateSegmentRequest
	segmentErr  error

	constraintReqs []*flipt.CreateConstraintRequest
	constraintErr  error

	ruleReqs []*flipt.CreateRuleRequest
	ruleErr  error

	distributionReqs []*flipt.CreateDistributionRequest
	distributionErr  error

	rolloutReqs []*flipt.CreateRolloutRequest
	rolloutErr  error

	listFlagReqs  []*flipt.ListFlagRequest
	listFlagResps []*flipt.FlagList
	listFlagErr   error

	listSegmentReqs  []*flipt.ListSegmentRequest
	listSegmentResps []*flipt.SegmentList
	listSegmentErr   error
}

func (m *mockCreator) GetNamespace(ctx context.Context, r *flipt.GetNamespaceRequest) (*flipt.Namespace, error) {
	m.getNSReqs = append(m.getNSReqs, r)
	return &flipt.Namespace{Key: "default"}, m.getNSErr
}

func (m *mockCreator) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	m.createNSReqs = append(m.createNSReqs, r)
	return &flipt.Namespace{Key: "default"}, m.createNSErr
}

func (m *mockCreator) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	m.createflagReqs = append(m.createflagReqs, r)
	if m.createflagErr != nil {
		return nil, m.createflagErr
	}
	return &flipt.Flag{
		NamespaceKey: r.NamespaceKey,
		Key:          r.Key,
		Name:         r.Name,
		Description:  r.Description,
		Type:         r.Type,
		Enabled:      r.Enabled,
		Metadata:     r.Metadata,
	}, nil
}

func (m *mockCreator) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	m.updateFlagReqs = append(m.updateFlagReqs, r)
	if m.updateFlagErr != nil {
		return nil, m.updateFlagErr
	}
	return &flipt.Flag{
		NamespaceKey: r.NamespaceKey,
		Key:          r.Key,
		Name:         r.Name,
		Description:  r.Description,
		DefaultVariant: &flipt.Variant{
			Id: r.DefaultVariantId,
		},
		Enabled: r.Enabled,
	}, nil
}

func (m *mockCreator) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	m.variantReqs = append(m.variantReqs, r)
	if m.variantErr != nil {
		return nil, m.variantErr
	}
	return &flipt.Variant{
		Id:          "static_variant_id",
		FlagKey:     r.FlagKey,
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
		Attachment:  r.Attachment,
	}, nil
}

func (m *mockCreator) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	m.segmentReqs = append(m.segmentReqs, r)
	if m.segmentErr != nil {
		return nil, m.segmentErr
	}
	return &flipt.Segment{
		Key:         r.Key,
		Name:        r.Name,
		Description: r.Description,
		MatchType:   r.MatchType,
	}, nil
}

func (m *mockCreator) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	m.constraintReqs = append(m.constraintReqs, r)
	if m.constraintErr != nil {
		return nil, m.constraintErr
	}
	return &flipt.Constraint{
		Id:         "static_constraint_id",
		SegmentKey: r.SegmentKey,
		Type:       r.Type,
		Property:   r.Property,
		Operator:   r.Operator,
		Value:      r.Value,
	}, nil
}

func (m *mockCreator) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	m.ruleReqs = append(m.ruleReqs, r)
	if m.ruleErr != nil {
		return nil, m.ruleErr
	}
	return &flipt.Rule{
		Id:         "static_rule_id",
		FlagKey:    r.FlagKey,
		SegmentKey: r.SegmentKey,
		Rank:       r.Rank,
	}, nil
}

func (m *mockCreator) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	m.distributionReqs = append(m.distributionReqs, r)
	if m.distributionErr != nil {
		return nil, m.distributionErr
	}
	return &flipt.Distribution{
		Id:        "static_distribution_id",
		RuleId:    r.RuleId,
		VariantId: r.VariantId,
		Rollout:   r.Rollout,
	}, nil
}

func (m *mockCreator) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	m.rolloutReqs = append(m.rolloutReqs, r)
	if m.rolloutErr != nil {
		return nil, m.rolloutErr
	}

	rollout := &flipt.Rollout{
		Id:           "static_rollout_id",
		NamespaceKey: r.NamespaceKey,
		FlagKey:      r.FlagKey,
		Description:  r.Description,
		Rank:         r.Rank,
	}

	switch rule := r.Rule.(type) {
	case *flipt.CreateRolloutRequest_Threshold:
		rollout.Rule = &flipt.Rollout_Threshold{
			Threshold: rule.Threshold,
		}
	case *flipt.CreateRolloutRequest_Segment:
		rollout.Rule = &flipt.Rollout_Segment{
			Segment: rule.Segment,
		}
	default:
		return nil, errors.New("unexpected rollout rule type")
	}

	return rollout, nil
}

func (m *mockCreator) ListFlags(ctx context.Context, r *flipt.ListFlagRequest) (*flipt.FlagList, error) {
	m.listFlagReqs = append(m.listFlagReqs, r)
	if m.listFlagErr != nil {
		return nil, m.listFlagErr
	}

	if len(m.listFlagResps) == 0 {
		return nil, fmt.Errorf("no response for ListFlags request: %+v", r)
	}

	var resp *flipt.FlagList
	resp, m.listFlagResps = m.listFlagResps[0], m.listFlagResps[1:]
	return resp, nil
}

func (m *mockCreator) ListSegments(ctx context.Context, r *flipt.ListSegmentRequest) (*flipt.SegmentList, error) {
	m.listSegmentReqs = append(m.listSegmentReqs, r)
	if m.listSegmentErr != nil {
		return nil, m.listSegmentErr
	}

	if len(m.listSegmentResps) == 0 {
		return nil, fmt.Errorf("no response for ListSegments request: %+v", r)
	}

	var resp *flipt.SegmentList
	resp, m.listSegmentResps = m.listSegmentResps[0], m.listSegmentResps[1:]
	return resp, nil
}

const variantAttachment = `{
  "pi": 3.141,
  "happy": true,
  "name": "Niels",
  "answer": {
    "everything": 42
  },
  "list": [1, 0, 2],
  "object": {
    "currency": "USD",
    "value": 42.99
  }
}`

func TestImport(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		skipExisting bool
		creator      func() *mockCreator
		expected     *mockCreator
	}{
		{
			name: "import with attachment and default variant",
			path: "testdata/import",
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag1",
						Name:        "flag1",
						Description: "description",
						Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
						Enabled:     true,
					},
					{
						Key:         "flag2",
						Name:        "flag2",
						Description: "a boolean flag",
						Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
						Enabled:     false,
					},
				},
				variantReqs: []*flipt.CreateVariantRequest{
					{
						FlagKey:     "flag1",
						Key:         "variant1",
						Name:        "variant1",
						Description: "variant description",
						Attachment:  compact(t, variantAttachment),
					},
				},
				updateFlagReqs: []*flipt.UpdateFlagRequest{
					{
						Key:              "flag1",
						Name:             "flag1",
						Description:      "description",
						Enabled:          true,
						DefaultVariantId: "static_variant_id",
					},
				},
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment1",
						Name:        "segment1",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment1",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "fizz",
						Operator:   "neq",
						Value:      "buzz",
					},
				},
				ruleReqs: []*flipt.CreateRuleRequest{
					{
						FlagKey:    "flag1",
						SegmentKey: "segment1",
						Rank:       1,
					},
				},
				distributionReqs: []*flipt.CreateDistributionRequest{
					{
						RuleId:    "static_rule_id",
						VariantId: "static_variant_id",
						FlagKey:   "flag1",
						Rollout:   100,
					},
				},
				rolloutReqs: []*flipt.CreateRolloutRequest{
					{
						FlagKey:     "flag2",
						Description: "enabled for internal users",
						Rank:        1,
						Rule: &flipt.CreateRolloutRequest_Segment{
							Segment: &flipt.RolloutSegment{
								SegmentKey: "internal_users",
								Value:      true,
							},
						},
					},
					{
						FlagKey:     "flag2",
						Description: "enabled for 50%",
						Rank:        2,
						Rule: &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: 50.0,
								Value:      true,
							},
						},
					},
				},
			},
		},
		{
			name: "import with attachment",
			path: "testdata/import_with_attachment",
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag1",
						Name:        "flag1",
						Description: "description",
						Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
						Enabled:     true,
					},
					{
						Key:         "flag2",
						Name:        "flag2",
						Description: "a boolean flag",
						Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
						Enabled:     false,
					},
				},
				variantReqs: []*flipt.CreateVariantRequest{
					{
						FlagKey:     "flag1",
						Key:         "variant1",
						Name:        "variant1",
						Description: "variant description",
						Attachment:  compact(t, variantAttachment),
					},
				},
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment1",
						Name:        "segment1",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment1",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "fizz",
						Operator:   "neq",
						Value:      "buzz",
					},
				},
				ruleReqs: []*flipt.CreateRuleRequest{
					{
						FlagKey:    "flag1",
						SegmentKey: "segment1",
						Rank:       1,
					},
				},
				distributionReqs: []*flipt.CreateDistributionRequest{
					{
						RuleId:    "static_rule_id",
						VariantId: "static_variant_id",
						FlagKey:   "flag1",
						Rollout:   100,
					},
				},
				rolloutReqs: []*flipt.CreateRolloutRequest{
					{
						FlagKey:     "flag2",
						Description: "enabled for internal users",
						Rank:        1,
						Rule: &flipt.CreateRolloutRequest_Segment{
							Segment: &flipt.RolloutSegment{
								SegmentKey: "internal_users",
								Value:      true,
							},
						},
					},
					{
						FlagKey:     "flag2",
						Description: "enabled for 50%",
						Rank:        2,
						Rule: &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: 50.0,
								Value:      true,
							},
						},
					},
				},
			},
		},
		{
			name: "import without attachment",
			path: "testdata/import_no_attachment",
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag1",
						Name:        "flag1",
						Description: "description",
						Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
						Enabled:     true,
					},
					{
						Key:         "flag2",
						Name:        "flag2",
						Description: "a boolean flag",
						Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
						Enabled:     false,
					},
				},
				variantReqs: []*flipt.CreateVariantRequest{
					{
						FlagKey:     "flag1",
						Key:         "variant1",
						Name:        "variant1",
						Description: "variant description",
					},
				},
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment1",
						Name:        "segment1",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment1",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "fizz",
						Operator:   "neq",
						Value:      "buzz",
					},
				},
				ruleReqs: []*flipt.CreateRuleRequest{
					{
						FlagKey:    "flag1",
						SegmentKey: "segment1",
						Rank:       1,
					},
				},
				distributionReqs: []*flipt.CreateDistributionRequest{
					{
						RuleId:    "static_rule_id",
						VariantId: "static_variant_id",
						FlagKey:   "flag1",
						Rollout:   100,
					},
				},
				rolloutReqs: []*flipt.CreateRolloutRequest{
					{
						FlagKey:     "flag2",
						Description: "enabled for internal users",
						Rank:        1,
						Rule: &flipt.CreateRolloutRequest_Segment{
							Segment: &flipt.RolloutSegment{
								SegmentKey: "internal_users",
								Value:      true,
							},
						},
					},
					{
						FlagKey:     "flag2",
						Description: "enabled for 50%",
						Rank:        2,
						Rule: &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: 50.0,
								Value:      true,
							},
						},
					},
				},
			},
		},
		{
			name: "import with implicit rule ranks",
			path: "testdata/import_implicit_rule_rank",
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag1",
						Name:        "flag1",
						Description: "description",
						Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
						Enabled:     true,
					},
					{
						Key:         "flag2",
						Name:        "flag2",
						Description: "a boolean flag",
						Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
						Enabled:     false,
					},
				},
				variantReqs: []*flipt.CreateVariantRequest{
					{
						FlagKey:     "flag1",
						Key:         "variant1",
						Name:        "variant1",
						Description: "variant description",
						Attachment:  compact(t, variantAttachment),
					},
				},
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment1",
						Name:        "segment1",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment1",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "fizz",
						Operator:   "neq",
						Value:      "buzz",
					},
				},
				ruleReqs: []*flipt.CreateRuleRequest{
					{
						FlagKey:    "flag1",
						SegmentKey: "segment1",
						Rank:       1,
					},
				},
				distributionReqs: []*flipt.CreateDistributionRequest{
					{
						RuleId:    "static_rule_id",
						VariantId: "static_variant_id",
						FlagKey:   "flag1",
						Rollout:   100,
					},
				},
				rolloutReqs: []*flipt.CreateRolloutRequest{
					{
						FlagKey:     "flag2",
						Description: "enabled for internal users",
						Rank:        1,
						Rule: &flipt.CreateRolloutRequest_Segment{
							Segment: &flipt.RolloutSegment{
								SegmentKey: "internal_users",
								Value:      true,
							},
						},
					},
					{
						FlagKey:     "flag2",
						Description: "enabled for 50%",
						Rank:        2,
						Rule: &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: 50.0,
								Value:      true,
							},
						},
					},
				},
			},
		},
		{
			name: "import with multiple segments",
			path: "testdata/import_rule_multiple_segments",
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag1",
						Name:        "flag1",
						Description: "description",
						Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
						Enabled:     true,
					},
					{
						Key:         "flag2",
						Name:        "flag2",
						Description: "a boolean flag",
						Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
						Enabled:     false,
					},
				},
				variantReqs: []*flipt.CreateVariantRequest{
					{
						FlagKey:     "flag1",
						Key:         "variant1",
						Name:        "variant1",
						Description: "variant description",
						Attachment:  compact(t, variantAttachment),
					},
				},
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment1",
						Name:        "segment1",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment1",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "fizz",
						Operator:   "neq",
						Value:      "buzz",
					},
				},
				ruleReqs: []*flipt.CreateRuleRequest{
					{
						FlagKey:     "flag1",
						SegmentKeys: []string{"segment1"},
						Rank:        1,
					},
				},
				distributionReqs: []*flipt.CreateDistributionRequest{
					{
						RuleId:    "static_rule_id",
						VariantId: "static_variant_id",
						FlagKey:   "flag1",
						Rollout:   100,
					},
				},
				rolloutReqs: []*flipt.CreateRolloutRequest{
					{
						FlagKey:     "flag2",
						Description: "enabled for internal users",
						Rank:        1,
						Rule: &flipt.CreateRolloutRequest_Segment{
							Segment: &flipt.RolloutSegment{
								SegmentKey: "internal_users",
								Value:      true,
							},
						},
					},
					{
						FlagKey:     "flag2",
						Description: "enabled for 50%",
						Rank:        2,
						Rule: &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: 50.0,
								Value:      true,
							},
						},
					},
				},
			},
		},
		{
			name: "import v1",
			path: "testdata/import_v1",
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag1",
						Name:        "flag1",
						Description: "description",
						Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
						Enabled:     true,
					},
				},
				variantReqs: []*flipt.CreateVariantRequest{
					{
						FlagKey:     "flag1",
						Key:         "variant1",
						Name:        "variant1",
						Description: "variant description",
						Attachment:  compact(t, variantAttachment),
					},
				},
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment1",
						Name:        "segment1",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment1",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "fizz",
						Operator:   "neq",
						Value:      "buzz",
					},
				},
				ruleReqs: []*flipt.CreateRuleRequest{
					{
						FlagKey:    "flag1",
						SegmentKey: "segment1",
						Rank:       1,
					},
				},
				distributionReqs: []*flipt.CreateDistributionRequest{
					{
						RuleId:    "static_rule_id",
						VariantId: "static_variant_id",
						FlagKey:   "flag1",
						Rollout:   100,
					},
				},
			},
		},
		{
			name: "import v1.1",
			path: "testdata/import_v1_1",
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag1",
						Name:        "flag1",
						Description: "description",
						Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
						Enabled:     true,
					},
					{
						Key:         "flag2",
						Name:        "flag2",
						Description: "a boolean flag",
						Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
						Enabled:     false,
					},
				},
				variantReqs: []*flipt.CreateVariantRequest{
					{
						FlagKey:     "flag1",
						Key:         "variant1",
						Name:        "variant1",
						Description: "variant description",
						Attachment:  compact(t, variantAttachment),
					},
				},
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment1",
						Name:        "segment1",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment1",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "fizz",
						Operator:   "neq",
						Value:      "buzz",
					},
				},
				ruleReqs: []*flipt.CreateRuleRequest{
					{
						FlagKey:    "flag1",
						SegmentKey: "segment1",
						Rank:       1,
					},
				},
				distributionReqs: []*flipt.CreateDistributionRequest{
					{
						RuleId:    "static_rule_id",
						VariantId: "static_variant_id",
						FlagKey:   "flag1",
						Rollout:   100,
					},
				},
				rolloutReqs: []*flipt.CreateRolloutRequest{
					{
						FlagKey:     "flag2",
						Description: "enabled for internal users",
						Rank:        1,
						Rule: &flipt.CreateRolloutRequest_Segment{
							Segment: &flipt.RolloutSegment{
								SegmentKey: "internal_users",
								Value:      true,
							},
						},
					},
					{
						FlagKey:     "flag2",
						Description: "enabled for 50%",
						Rank:        2,
						Rule: &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: 50.0,
								Value:      true,
							},
						},
					},
				},
			},
		},
		{
			name:         "import new flags only",
			path:         "testdata/import_new_flags_only",
			skipExisting: true,
			creator: func() *mockCreator {
				return &mockCreator{
					listFlagResps: []*flipt.FlagList{{
						Flags: []*flipt.Flag{{
							Key:         "flag1",
							Name:        "flag1",
							Description: "description",
							Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
							Enabled:     true,
						}},
					}},
					listSegmentResps: []*flipt.SegmentList{{
						Segments: []*flipt.Segment{{
							Key:         "segment1",
							Name:        "segment1",
							Description: "description",
							MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
						}},
					}},
				}
			},
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag2",
						Name:        "flag2",
						Description: "a boolean flag",
						Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
						Enabled:     false,
					},
				},
				variantReqs: nil,
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment2",
						Name:        "segment2",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment2",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "buzz",
						Operator:   "neq",
						Value:      "fizz",
					},
				},
				ruleReqs:         nil,
				distributionReqs: nil,
				rolloutReqs: []*flipt.CreateRolloutRequest{
					{
						FlagKey:     "flag2",
						Description: "enabled for internal users",
						Rank:        1,
						Rule: &flipt.CreateRolloutRequest_Segment{
							Segment: &flipt.RolloutSegment{
								SegmentKey: "internal_users",
								Value:      true,
							},
						},
					},
					{
						FlagKey:     "flag2",
						Description: "enabled for 50%",
						Rank:        2,
						Rule: &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: 50.0,
								Value:      true,
							},
						},
					},
				},
				listFlagReqs: []*flipt.ListFlagRequest{
					{},
				},
				listFlagResps: []*flipt.FlagList{},
				listSegmentReqs: []*flipt.ListSegmentRequest{
					{},
				},
				listSegmentResps: []*flipt.SegmentList{},
			},
		},
		{
			name: "import v1.3",
			path: "testdata/import_v1_3",
			expected: &mockCreator{
				createflagReqs: []*flipt.CreateFlagRequest{
					{
						Key:         "flag1",
						Name:        "flag1",
						Description: "description",
						Type:        flipt.FlagType_VARIANT_FLAG_TYPE,
						Enabled:     true,
						Metadata:    newStruct(t, map[string]any{"label": "variant", "area": true}),
					},
					{
						Key:         "flag2",
						Name:        "flag2",
						Description: "a boolean flag",
						Type:        flipt.FlagType_BOOLEAN_FLAG_TYPE,
						Enabled:     false,
						Metadata:    newStruct(t, map[string]any{"label": "bool", "area": 12}),
					},
				},
				variantReqs: []*flipt.CreateVariantRequest{
					{
						FlagKey:     "flag1",
						Key:         "variant1",
						Name:        "variant1",
						Description: "variant description",
						Attachment:  compact(t, variantAttachment),
					},
				},
				segmentReqs: []*flipt.CreateSegmentRequest{
					{
						Key:         "segment1",
						Name:        "segment1",
						Description: "description",
						MatchType:   flipt.MatchType_ANY_MATCH_TYPE,
					},
				},
				constraintReqs: []*flipt.CreateConstraintRequest{
					{
						SegmentKey: "segment1",
						Type:       flipt.ComparisonType_STRING_COMPARISON_TYPE,
						Property:   "fizz",
						Operator:   "neq",
						Value:      "buzz",
					},
				},
				ruleReqs: []*flipt.CreateRuleRequest{
					{
						FlagKey:    "flag1",
						SegmentKey: "segment1",
						Rank:       1,
					},
				},
				distributionReqs: []*flipt.CreateDistributionRequest{
					{
						RuleId:    "static_rule_id",
						VariantId: "static_variant_id",
						FlagKey:   "flag1",
						Rollout:   100,
					},
				},
				rolloutReqs: []*flipt.CreateRolloutRequest{
					{
						FlagKey:     "flag2",
						Description: "enabled for internal users",
						Rank:        1,
						Rule: &flipt.CreateRolloutRequest_Segment{
							Segment: &flipt.RolloutSegment{
								SegmentKey: "internal_users",
								Value:      true,
							},
						},
					},
					{
						FlagKey:     "flag2",
						Description: "enabled for 50%",
						Rank:        2,
						Rule: &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: 50.0,
								Value:      true,
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		for _, ext := range extensions {
			t.Run(fmt.Sprintf("%s (%s)", tc.name, ext), func(t *testing.T) {
				creator := &mockCreator{}
				if tc.creator != nil {
					creator = tc.creator()
				}
				importer := NewImporter(creator)

				in, err := os.Open(tc.path + "." + string(ext))
				assert.NoError(t, err)
				defer in.Close()

				err = importer.Import(context.Background(), ext, in, tc.skipExisting)
				assert.NoError(t, err)

				assert.Equal(t, tc.expected, creator)
			})
		}
	}
}

func TestImport_Export(t *testing.T) {
	var (
		creator  = &mockCreator{}
		importer = NewImporter(creator)
	)

	in, err := os.Open("testdata/export.yml")
	assert.NoError(t, err)
	defer in.Close()

	err = importer.Import(context.Background(), EncodingYML, in, skipExistingFalse)
	require.NoError(t, err)
	assert.Equal(t, "default", creator.createflagReqs[0].NamespaceKey)
}

func TestImport_InvalidVersion(t *testing.T) {
	var (
		creator  = &mockCreator{}
		importer = NewImporter(creator)
	)

	for _, ext := range extensions {
		in, err := os.Open("testdata/import_invalid_version." + string(ext))
		assert.NoError(t, err)
		defer in.Close()

		err = importer.Import(context.Background(), ext, in, skipExistingFalse)
		assert.EqualError(t, err, "unsupported version: 5.0")
	}
}

func TestImport_FlagType_LTVersion1_1(t *testing.T) {
	var (
		creator  = &mockCreator{}
		importer = NewImporter(creator)
	)

	for _, ext := range extensions {
		in, err := os.Open("testdata/import_v1_flag_type_not_supported." + string(ext))
		assert.NoError(t, err)
		defer in.Close()

		err = importer.Import(context.Background(), ext, in, skipExistingFalse)
		assert.EqualError(t, err, "flag.type is supported in version >=1.1, found 1.0")
	}
}

func TestImport_Rollouts_LTVersion1_1(t *testing.T) {
	var (
		creator  = &mockCreator{}
		importer = NewImporter(creator)
	)

	for _, ext := range extensions {
		in, err := os.Open("testdata/import_v1_rollouts_not_supported." + string(ext))
		assert.NoError(t, err)
		defer in.Close()

		err = importer.Import(context.Background(), ext, in, skipExistingFalse)
		assert.EqualError(t, err, "flag.rollouts is supported in version >=1.1, found 1.0")
	}
}

func TestImport_Namespaces_Mix_And_Match(t *testing.T) {
	tests := []struct {
		name                      string
		path                      string
		expectedGetNSReqs         int
		expectedCreateFlagReqs    int
		expectedCreateSegmentReqs int
	}{
		{
			name:                      "single namespace no YAML stream",
			path:                      "testdata/import",
			expectedGetNSReqs:         0,
			expectedCreateFlagReqs:    2,
			expectedCreateSegmentReqs: 1,
		},
		{
			name:                      "single namespace foo non-YAML stream",
			path:                      "testdata/import_single_namespace_foo",
			expectedGetNSReqs:         1,
			expectedCreateFlagReqs:    2,
			expectedCreateSegmentReqs: 1,
		},
		{
			name:                      "multiple namespaces default and foo",
			path:                      "testdata/import_two_namespaces_default_and_foo",
			expectedGetNSReqs:         1,
			expectedCreateFlagReqs:    4,
			expectedCreateSegmentReqs: 2,
		},
		{
			name:                      "yaml stream only default namespace",
			path:                      "testdata/import_yaml_stream_default_namespace",
			expectedGetNSReqs:         0,
			expectedCreateFlagReqs:    4,
			expectedCreateSegmentReqs: 2,
		},
		{
			name:                      "yaml stream all unqiue namespaces",
			path:                      "testdata/import_yaml_stream_all_unique_namespaces",
			expectedGetNSReqs:         2,
			expectedCreateFlagReqs:    6,
			expectedCreateSegmentReqs: 3,
		},
	}

	for _, tc := range tests {
		tc := tc
		for _, ext := range extensions {
			t.Run(fmt.Sprintf("%s (%s)", tc.name, ext), func(t *testing.T) {
				var (
					creator  = &mockCreator{}
					importer = NewImporter(creator)
				)

				in, err := os.Open(tc.path + "." + string(ext))
				assert.NoError(t, err)
				defer in.Close()

				err = importer.Import(context.Background(), ext, in, skipExistingFalse)
				assert.NoError(t, err)

				assert.Len(t, creator.getNSReqs, tc.expectedGetNSReqs)
				assert.Len(t, creator.createflagReqs, tc.expectedCreateFlagReqs)
				assert.Len(t, creator.segmentReqs, tc.expectedCreateSegmentReqs)
			})
		}
	}
}

//nolint:unparam
func compact(t *testing.T, v string) string {
	t.Helper()

	var m any
	require.NoError(t, json.Unmarshal([]byte(v), &m))

	d, err := json.Marshal(m)
	require.NoError(t, err)

	return string(d)
}
