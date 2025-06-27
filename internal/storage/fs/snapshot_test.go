package fs

import (
	"context"
	"embed"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

//go:embed testdata
var testdata embed.FS

func ptr[P any](p P) *P {
	return &p
}

func TestSnapshot_GetFlag(t *testing.T) {
	conf := DefaultFliptConfig()
	snap, err := SnapshotFromFS(zaptest.NewLogger(t), conf, testdata)
	require.NoError(t, err)

	tests := []struct {
		name      string
		namespace string
		key       string
		wantFlag  *core.Flag
		wantErr   error
	}{
		{
			name:      "existing flag",
			namespace: "production",
			key:       "prod-flag-1",
			wantFlag: &core.Flag{
				Key:            "prod-flag-1",
				Name:           "Prod Flag 1",
				Description:    "description",
				Enabled:        true,
				DefaultVariant: ptr("prod-variant"),
				Variants: []*core.Variant{
					{
						Key:  "prod-variant",
						Name: "Prod Variant",
						Attachment: func() *structpb.Value {
							v, _ := structpb.NewValue(map[string]any{
								"pi":      3.141,
								"happy":   true,
								"name":    "Niels",
								"nothing": nil,
								"answer": map[string]any{
									"everything": 42,
								},
								"list": []any{1, 0, 2},
								"object": map[string]any{
									"currency": "USD",
									"value":    42.99,
								},
							})
							return v
						}(),
					},
					{
						Key:        "foo",
						Attachment: structpb.NewNullValue(),
					},
				},
				Rules: []*core.Rule{
					{
						Segments: []string{"segment2"},
						Distributions: []*core.Distribution{
							{
								Variant: "prod-variant",
								Rollout: 100,
							},
						},
					},
				},
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"foo":    {Kind: &structpb.Value_StringValue{StringValue: "bar"}},
						"number": {Kind: &structpb.Value_NumberValue{NumberValue: 42}},
					},
				},
				Type: core.FlagType_VARIANT_FLAG_TYPE,
			},
		},
		{
			name:      "boolean flag",
			namespace: "production",
			key:       "flag_boolean_2",
			wantFlag: &core.Flag{
				Key:         "flag_boolean_2",
				Name:        "FLAG_BOOLEAN",
				Type:        core.FlagType_BOOLEAN_FLAG_TYPE,
				Description: "Boolean Flag Description",
				Enabled:     false,
				Rollouts: []*core.Rollout{
					{
						Description: "enabled for segment2",
						Type:        core.RolloutType_SEGMENT_ROLLOUT_TYPE,
						Rule: &core.Rollout_Segment{
							Segment: &core.RolloutSegment{
								Value:    true,
								Segments: []string{"segment2"},
							},
						},
					},
					{
						Description: "enabled for 50%",
						Type:        core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
						Rule: &core.Rollout_Threshold{
							Threshold: &core.RolloutThreshold{
								Percentage: 50,
								Value:      true,
							},
						},
					},
				},
			},
		},
		{
			name:      "non-existent flag",
			namespace: "production",
			key:       "non-existent",
			wantErr:   errors.ErrNotFound(`flag "production/non-existent"`),
		},
		{
			name:      "non-existent namespace",
			namespace: "non-existent",
			key:       "prod-flag-1",
			wantErr:   errors.ErrNotFound(`namespace "non-existent"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag, err := snap.GetFlag(context.Background(), storage.NewResource(
				tt.namespace,
				tt.key,
			))

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			if diff := cmp.Diff(tt.wantFlag, flag, protocmp.Transform()); diff != "" {
				t.Errorf("GetFlag() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSnapshot_ListFlags(t *testing.T) {
	snap, err := SnapshotFromFS(zaptest.NewLogger(t), DefaultFliptConfig(), testdata)
	require.NoError(t, err)

	tests := []struct {
		name      string
		namespace string
		wantFlags []*core.Flag
		wantErr   error
	}{
		{
			name:      "production namespace flags",
			namespace: "production",
			wantFlags: []*core.Flag{
				{
					Key:         "flag_boolean_2",
					Name:        "FLAG_BOOLEAN",
					Type:        core.FlagType_BOOLEAN_FLAG_TYPE,
					Description: "Boolean Flag Description",
					Enabled:     false,
					Rollouts: []*core.Rollout{
						{
							Type:        core.RolloutType_SEGMENT_ROLLOUT_TYPE,
							Description: "enabled for segment2",
							Rule: &core.Rollout_Segment{
								Segment: &core.RolloutSegment{
									Value:    true,
									Segments: []string{"segment2"},
								},
							},
						},
						{
							Type:        core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
							Description: "enabled for 50%",
							Rule: &core.Rollout_Threshold{
								Threshold: &core.RolloutThreshold{
									Percentage: 50,
									Value:      true,
								},
							},
						},
					},
				},
				{
					Key:            "prod-flag-1",
					Name:           "Prod Flag 1",
					Description:    "description",
					Enabled:        true,
					DefaultVariant: ptr("prod-variant"),
					Variants: []*core.Variant{
						{
							Key:  "prod-variant",
							Name: "Prod Variant",
							Attachment: func() *structpb.Value {
								v, _ := structpb.NewValue(map[string]any{
									"pi":      3.141,
									"happy":   true,
									"name":    "Niels",
									"nothing": nil,
									"answer": map[string]any{
										"everything": 42,
									},
									"list": []any{1, 0, 2},
									"object": map[string]any{
										"currency": "USD",
										"value":    42.99,
									},
								})
								return v
							}(),
						},
						{
							Key:        "foo",
							Attachment: structpb.NewNullValue(),
						},
					},
					Rules: []*core.Rule{
						{
							Segments: []string{"segment2"},
							Distributions: []*core.Distribution{
								{
									Variant: "prod-variant",
									Rollout: 100,
								},
							},
						},
					},
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"foo":    {Kind: &structpb.Value_StringValue{StringValue: "bar"}},
							"number": {Kind: &structpb.Value_NumberValue{NumberValue: 42}},
						},
					},
					Type: core.FlagType_VARIANT_FLAG_TYPE,
				},
			},
		},
		{
			name:      "empty namespace returns empty list",
			namespace: "empty",
			wantFlags: nil,
		},
		{
			name:      "non-existent namespace",
			namespace: "non-existent",
			wantErr:   errors.ErrNotFoundf("namespace %q", "non-existent"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, err := snap.ListFlags(context.Background(), &storage.ListRequest[storage.NamespaceRequest]{
				Predicate:   storage.NewNamespace(tt.namespace),
				QueryParams: storage.QueryParams{},
			})

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.NoError(t, err)

			opts := []cmp.Option{
				protocmp.Transform(),
				protocmp.SortRepeated(func(a, b *core.Flag) bool {
					return a.Key < b.Key
				}),
			}

			if !cmp.Equal(tt.wantFlags, set.Results, opts...) {
				t.Errorf("flags differ: %v", cmp.Diff(tt.wantFlags, set.Results, opts...))
			}
		})
	}
}

func TestSnapshot_CountFlags(t *testing.T) {
	snap, err := SnapshotFromFS(zaptest.NewLogger(t), DefaultFliptConfig(), testdata)
	require.NoError(t, err)

	tests := []struct {
		name      string
		namespace string
		want      uint64
		wantErr   error
	}{
		{
			name:      "production namespace flags",
			namespace: "production",
			want:      2, // prod-flag-1 and flag_boolean_2
		},
		{
			name:      "empty namespace",
			namespace: "empty",
			want:      0,
		},
		{
			name:      "non-existent namespace",
			namespace: "non-existent",
			wantErr:   errors.ErrNotFoundf("namespace %q", "non-existent"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, err := snap.CountFlags(context.Background(), storage.NewNamespace(tt.namespace))

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, count)
		})
	}
}

func TestSnapshot_GetEvaluationRules(t *testing.T) {
	snap, err := SnapshotFromFS(zaptest.NewLogger(t), DefaultFliptConfig(), testdata)
	require.NoError(t, err)

	tests := []struct {
		name      string
		namespace string
		flagKey   string
		want      []*storage.EvaluationRule
		wantErr   error
	}{
		{
			name:      "variant flag rules",
			namespace: "production",
			flagKey:   "prod-flag-1",
			want: []*storage.EvaluationRule{
				{
					ID:           "1", // Note: ID is generated with UUID, use mock or ignore in comparison
					FlagKey:      "prod-flag-1",
					NamespaceKey: "production",
					Rank:         1,
					Segments: map[string]*storage.EvaluationSegment{
						"segment2": {
							SegmentKey: "segment2",
							MatchType:  core.MatchType_ANY_MATCH_TYPE,
							Constraints: []storage.EvaluationConstraint{
								{
									Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
									Property: "foo",
									Operator: "eq",
									Value:    "baz",
								},
								{
									Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
									Property: "fizz",
									Operator: "neq",
									Value:    "buzz",
								},
							},
						},
					},
				},
			},
		},
		{
			name:      "boolean flag rules",
			namespace: "production",
			flagKey:   "flag_boolean_2",
			want:      []*storage.EvaluationRule{}, // Boolean flags use rollouts instead of rules
		},
		{
			name:      "non-existent flag",
			namespace: "production",
			flagKey:   "non-existent",
			wantErr:   errors.ErrNotFoundf("flag %q", "production/non-existent"),
		},
		{
			name:      "non-existent namespace",
			namespace: "non-existent",
			flagKey:   "prod-flag-1",
			wantErr:   errors.ErrNotFoundf("namespace %q", "non-existent"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := snap.GetEvaluationRules(context.Background(), storage.NewResource(
				tt.namespace,
				tt.flagKey,
			))

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.NoError(t, err)

			// Since rule IDs are generated with UUIDs, we need to ignore them in comparison
			for i, rule := range rules {
				if len(tt.want) > i {
					tt.want[i].ID = rule.ID
				}
			}

			opts := []cmp.Option{
				cmp.AllowUnexported(storage.EvaluationConstraint{}),
				protocmp.Transform(),
			}

			if !cmp.Equal(tt.want, rules, opts...) {
				t.Errorf("rules differ: %v", cmp.Diff(tt.want, rules, opts...))
			}
		})
	}
}

type EvaluationDistribution struct {
	ID                string
	RuleID            string
	VariantID         string
	Rollout           float32
	VariantKey        string
	VariantAttachment map[string]any
}

// storageEvaluationDistTransformer allows us to ensure the contents of the attachment
// is preserved as JSON. We cannot rely on protojson output to be stable
// as it is unstable by design.
func storageEvaluationDistTransformer() cmp.Option {
	return cmp.Transformer("storage.EvaluationDistribution", func(dist *storage.EvaluationDistribution) EvaluationDistribution {
		attachment := make(map[string]any)
		if err := json.Unmarshal([]byte(dist.VariantAttachment), &attachment); err != nil {
			panic(err)
		}

		return EvaluationDistribution{
			ID:                dist.ID,
			RuleID:            dist.RuleID,
			VariantID:         dist.VariantID,
			Rollout:           dist.Rollout,
			VariantKey:        dist.VariantKey,
			VariantAttachment: attachment,
		}
	})
}

func TestSnapshot_GetEvaluationDistributions(t *testing.T) {
	snap, err := SnapshotFromFS(zaptest.NewLogger(t), DefaultFliptConfig(), testdata)
	require.NoError(t, err)

	// First get the rules to get valid rule IDs
	rules, err := snap.GetEvaluationRules(context.Background(), storage.NewResource(
		"production",
		"prod-flag-1",
	))
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	ruleID := rules[0].ID

	tests := []struct {
		name      string
		namespace string
		flagKey   string
		ruleID    string
		want      []*storage.EvaluationDistribution
		wantErr   error
	}{
		{
			name:      "get distributions for variant flag rule",
			namespace: "production",
			flagKey:   "prod-flag-1",
			ruleID:    ruleID,
			want: []*storage.EvaluationDistribution{
				{
					ID:                "1", // Note: ID is generated with UUID, will be replaced in test
					VariantKey:        "prod-variant",
					VariantAttachment: `{"answer":{"everything":42},"happy":true,"list":[1,0,2],"name":"Niels","nothing":null,"object":{"currency":"USD","value":42.99},"pi":3.141}`,
					Rollout:           100,
				},
			},
		},
		{
			name:      "non-existent rule ID",
			namespace: "production",
			flagKey:   "prod-flag-1",
			ruleID:    "non-existent",
			want:      []*storage.EvaluationDistribution{}, // Empty slice for non-existent rule
		},
		{
			name:      "boolean flag has no distributions",
			namespace: "production",
			flagKey:   "flag_boolean_2",
			ruleID:    "any-id",
			want:      []*storage.EvaluationDistribution{},
		},
		{
			name:      "non-existent namespace",
			namespace: "non-existent",
			flagKey:   "prod-flag-1",
			ruleID:    "any-id",
			wantErr:   errors.ErrNotFoundf("namespace %q", "non-existent"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dists, err := snap.GetEvaluationDistributions(
				context.Background(),
				storage.NewResource(tt.namespace, tt.flagKey),
				storage.NewID(tt.ruleID),
			)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.NoError(t, err)

			// Since distribution IDs are generated with UUIDs, we need to ignore them in comparison
			for i, dist := range dists {
				if len(tt.want) > i {
					tt.want[i].ID = dist.ID
				}
			}

			opts := []cmp.Option{
				cmp.AllowUnexported(),
				protocmp.Transform(),
				storageEvaluationDistTransformer(),
			}

			if !cmp.Equal(tt.want, dists, opts...) {
				t.Errorf("distributions differ: %v", cmp.Diff(tt.want, dists, opts...))
			}
		})
	}
}

func TestSnapshot_GetEvaluationRollouts(t *testing.T) {
	snap, err := SnapshotFromFS(zaptest.NewLogger(t), DefaultFliptConfig(), testdata)
	require.NoError(t, err)

	tests := []struct {
		name      string
		namespace string
		flagKey   string
		want      []*storage.EvaluationRollout
		wantErr   error
	}{
		{
			name:      "boolean flag rollouts",
			namespace: "production",
			flagKey:   "flag_boolean_2",
			want: []*storage.EvaluationRollout{
				{
					NamespaceKey: "production",
					Rank:         1,
					RolloutType:  core.RolloutType_SEGMENT_ROLLOUT_TYPE,
					Segment: &storage.RolloutSegment{
						Value: true,
						Segments: map[string]*storage.EvaluationSegment{
							"segment2": {
								SegmentKey: "segment2",
								MatchType:  core.MatchType_ANY_MATCH_TYPE,
								Constraints: []storage.EvaluationConstraint{
									{
										Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
										Property: "foo",
										Operator: "eq",
										Value:    "baz",
									},
									{
										Type:     core.ComparisonType_STRING_COMPARISON_TYPE,
										Property: "fizz",
										Operator: "neq",
										Value:    "buzz",
									},
								},
							},
						},
					},
				},
				{
					NamespaceKey: "production",
					Rank:         2,
					RolloutType:  core.RolloutType_THRESHOLD_ROLLOUT_TYPE,
					Threshold: &storage.RolloutThreshold{
						Percentage: 50,
						Value:      true,
					},
				},
			},
		},
		{
			name:      "variant flag has no rollouts",
			namespace: "production",
			flagKey:   "prod-flag-1",
			want:      []*storage.EvaluationRollout{},
		},
		{
			name:      "non-existent flag",
			namespace: "production",
			flagKey:   "non-existent",
			wantErr:   errors.ErrNotFoundf("flag %q", "production/non-existent"),
		},
		{
			name:      "non-existent namespace",
			namespace: "non-existent",
			flagKey:   "flag_boolean_2",
			wantErr:   errors.ErrNotFoundf("namespace %q", "non-existent"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rollouts, err := snap.GetEvaluationRollouts(context.Background(), storage.NewResource(
				tt.namespace,
				tt.flagKey,
			))

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.NoError(t, err)

			opts := []cmp.Option{
				cmp.AllowUnexported(storage.EvaluationConstraint{}),
				protocmp.Transform(),
			}

			if !cmp.Equal(tt.want, rollouts, opts...) {
				t.Errorf("rollouts differ: %v", cmp.Diff(tt.want, rollouts, opts...))
			}
		})
	}
}

func TestSnapshot_EvaluationNamespaceSnapshot(t *testing.T) {
	snap, err := SnapshotFromFS(zaptest.NewLogger(t), DefaultFliptConfig(), testdata)
	require.NoError(t, err)

	tests := []struct {
		name      string
		namespace string
		wantSnap  *evaluation.EvaluationNamespaceSnapshot
		wantErr   error
	}{
		{
			name:      "production namespace",
			namespace: "production",
			wantSnap: &evaluation.EvaluationNamespaceSnapshot{
				Digest: "09da9bff218a2826e61325bcc9b72f20def6b096",
				Namespace: &evaluation.EvaluationNamespace{
					Key: "production",
				},
				Flags: []*evaluation.EvaluationFlag{
					{
						Key:         "flag_boolean_2",
						Name:        "FLAG_BOOLEAN",
						Description: "Boolean Flag Description",
						Enabled:     false,
						Type:        evaluation.EvaluationFlagType_BOOLEAN_FLAG_TYPE,
						Rollouts: []*evaluation.EvaluationRollout{
							{
								Type: evaluation.EvaluationRolloutType_SEGMENT_ROLLOUT_TYPE,
								Rank: 1,
								Rule: &evaluation.EvaluationRollout_Segment{
									Segment: &evaluation.EvaluationRolloutSegment{
										Value: true,
										Segments: []*evaluation.EvaluationSegment{
											{
												Key:         "segment2",
												Name:        "segment2",
												Description: "description",
												MatchType:   evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE,
												Constraints: []*evaluation.EvaluationConstraint{
													{
														Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
														Property: "foo",
														Operator: "eq",
														Value:    "baz",
													},
													{
														Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
														Property: "fizz",
														Operator: "neq",
														Value:    "buzz",
													},
												},
											},
										},
									},
								},
							},
							{
								Type: evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE,
								Rank: 2,
								Rule: &evaluation.EvaluationRollout_Threshold{
									Threshold: &evaluation.EvaluationRolloutThreshold{
										Percentage: 50,
										Value:      true,
									},
								},
							},
						},
					},
					{
						Key:         "prod-flag-1",
						Name:        "Prod Flag 1",
						Description: "description",
						Enabled:     true,
						Type:        evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE,
						DefaultVariant: &evaluation.EvaluationVariant{
							Key: "prod-variant",
						},
						Rules: []*evaluation.EvaluationRule{
							{
								Rank: 1,
								Segments: []*evaluation.EvaluationSegment{
									{
										Key:         "segment2",
										Name:        "segment2",
										Description: "description",
										MatchType:   evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE,
										Constraints: []*evaluation.EvaluationConstraint{
											{
												Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
												Property: "foo",
												Operator: "eq",
												Value:    "baz",
											},
											{
												Type:     evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE,
												Property: "fizz",
												Operator: "neq",
												Value:    "buzz",
											},
										},
									},
								},
								Distributions: []*evaluation.EvaluationDistribution{
									{
										VariantKey:        "prod-variant",
										Rollout:           100,
										VariantAttachment: `{"answer":{"everything":42},"happy":true,"list":[1,0,2],"name":"Niels","nothing":null,"object":{"currency":"USD","value":42.99},"pi":3.141}`,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:      "empty namespace",
			namespace: "empty",
			wantSnap: &evaluation.EvaluationNamespaceSnapshot{
				Digest: "cb2710c073be3f2e8c51a1f8474b9725f51b6522",
				Namespace: &evaluation.EvaluationNamespace{
					Key: "empty",
				},
				Flags: []*evaluation.EvaluationFlag{},
			},
		},
		{
			name:      "non-existent namespace",
			namespace: "non-existent",
			wantErr:   errors.ErrNotFoundf("evaluation snapshot for namespace %q", "non-existent"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snapshot, err := snap.EvaluationNamespaceSnapshot(context.Background(), tt.namespace)

			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}

			require.NoError(t, err)

			// Ignore timestamps in comparison since they're generated at runtime
			for _, flag := range snapshot.Flags {
				flag.CreatedAt = nil
				flag.UpdatedAt = nil
			}

			// Ignore rule IDs in comparison since they're generated
			for _, flag := range snapshot.Flags {
				for _, rule := range flag.Rules {
					rule.Id = ""
					for _, dist := range rule.Distributions {
						dist.RuleId = ""
					}
				}
			}

			// Ignore variant IDs in comparison since they're generated
			for _, flag := range snapshot.Flags {
				if flag.DefaultVariant != nil {
					flag.DefaultVariant.Id = ""
				}
			}

			opts := []cmp.Option{
				protocmp.Transform(),
				protocmp.SortRepeated(func(a, b *evaluation.EvaluationFlag) bool {
					return a.Key < b.Key
				}),
				protocmp.IgnoreFields(&evaluation.EvaluationVariant{}, "attachment"),
				protocmp.IgnoreFields(&evaluation.EvaluationDistribution{}, "variant_attachment"),
			}

			if !cmp.Equal(tt.wantSnap, snapshot, opts...) {
				t.Errorf("snapshot differs: %v", cmp.Diff(tt.wantSnap, snapshot, opts...))
			}
		})
	}
}
