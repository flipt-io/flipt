package fs

import (
	"context"
	"embed"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
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
							v, _ := structpb.NewValue(map[string]interface{}{
								"pi":      3.141,
								"happy":   true,
								"name":    "Niels",
								"nothing": nil,
								"answer": map[string]interface{}{
									"everything": 42,
								},
								"list": []interface{}{1, 0, 2},
								"object": map[string]interface{}{
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
								v, _ := structpb.NewValue(map[string]interface{}{
									"pi":      3.141,
									"happy":   true,
									"name":    "Niels",
									"nothing": nil,
									"answer": map[string]interface{}{
										"everything": 42,
									},
									"list": []interface{}{1, 0, 2},
									"object": map[string]interface{}{
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
