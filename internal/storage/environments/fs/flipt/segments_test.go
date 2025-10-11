package flipt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fstesting "go.flipt.io/flipt/internal/storage/environments/fs/testing"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap/zaptest"
)

var (
	segmentContents = `version: "1.5"
namespace:
  key: default
  name: Default
  description: The default namespace
segments:
  - key: segment1
    name: Segment 1
    description: A test segment
    match_type: ALL
    constraints:
      - type: STRING
        description: A test constraint
        property: country
        operator: eq
        value: US
`

	multiDocSegmentContents = `version: "1.5"
namespace:
  key: team_a
  name: Team A
  description: The Team A namespace
segments:
  - key: segment1
    name: Segment 1
    match_type: ALL
---
version: "1.5"
namespace:
  key: team_a
  name: Team A
  description: The Team A namespace
segments:
  - key: segment2
    name: Segment 2
    match_type: ANY
`
)

func TestSegmentStorage_GetResource(t *testing.T) {
	ctx := context.TODO()
	logger := zaptest.NewLogger(t)
	storage := NewSegmentStorage(logger)

	fs := fstesting.NewFilesystem(
		t,
		fstesting.WithDirectory(
			"default",
			fstesting.WithFile("features.yaml", segmentContents),
		),
		fstesting.WithDirectory(
			"team_a",
			fstesting.WithFile("features.yaml", multiDocSegmentContents),
		),
	)

	tests := []struct {
		name      string
		namespace string
		key       string
		wantErr   bool
	}{
		{
			name:      "get existing segment",
			namespace: "default",
			key:       "segment1",
		},
		{
			name:      "get segment from multi-doc",
			namespace: "team_a",
			key:       "segment2",
		},
		{
			name:      "segment not found",
			namespace: "default",
			key:       "nonexistent",
			wantErr:   true,
		},
		{
			name:      "namespace not found",
			namespace: "nonexistent",
			key:       "segment1",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := storage.GetResource(ctx, fs, tt.namespace, tt.key)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.namespace, resource.NamespaceKey)
			assert.Equal(t, tt.key, resource.Key)

			var segment core.Segment
			err = resource.Payload.UnmarshalTo(&segment)
			require.NoError(t, err)
			assert.Equal(t, tt.key, segment.Key)
		})
	}
}

func TestSegmentStorage_ListResources(t *testing.T) {
	ctx := context.TODO()
	logger := zaptest.NewLogger(t)
	storage := NewSegmentStorage(logger)

	fs := fstesting.NewFilesystem(
		t,
		fstesting.WithDirectory(
			"default",
			fstesting.WithFile("features.yaml", segmentContents),
		),
		fstesting.WithDirectory(
			"team_a",
			fstesting.WithFile("features.yaml", multiDocSegmentContents),
		),
	)

	tests := []struct {
		name      string
		namespace string
		want      int
		wantErr   bool
	}{
		{
			name:      "list segments in default namespace",
			namespace: "default",
			want:      1,
		},
		{
			name:      "list segments in multi-doc namespace",
			namespace: "team_a",
			want:      2,
		},
		{
			name:      "list segments in non-existent namespace",
			namespace: "nonexistent",
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources, err := storage.ListResources(ctx, fs, tt.namespace)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resources, tt.want)

			for _, resource := range resources {
				assert.Equal(t, tt.namespace, resource.NamespaceKey)
				var segment core.Segment
				err = resource.Payload.UnmarshalTo(&segment)
				require.NoError(t, err)
			}
		})
	}
}

func TestSegmentStorage_PutResource(t *testing.T) {
	ctx := context.TODO()
	logger := zaptest.NewLogger(t)
	storage := NewSegmentStorage(logger)

	t.Run("create new segment", func(t *testing.T) {
		fs := fstesting.NewFilesystem(
			t,
			fstesting.WithDirectory(
				"default",
				fstesting.WithFile("features.yaml", defaultContents),
			),
		)

		segment := &core.Segment{
			Key:         "new_segment",
			Name:        "New Segment",
			Description: "A new test segment",
			MatchType:   core.MatchType_ALL_MATCH_TYPE,
			Constraints: []*core.Constraint{
				{
					Type:        core.ComparisonType_STRING_COMPARISON_TYPE,
					Description: "A test constraint",
					Property:    "country",
					Operator:    "eq",
					Value:       "US",
				},
			},
		}

		any, err := newAny(segment)
		require.NoError(t, err)

		err = storage.PutResource(ctx, fs, &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          segment.Key,
			Payload:      any,
		})
		require.NoError(t, err)

		// verify segment was created
		resource, err := storage.GetResource(ctx, fs, "default", segment.Key)
		require.NoError(t, err)
		assert.Equal(t, segment.Key, resource.Key)
	})

	t.Run("update existing segment", func(t *testing.T) {
		fs := fstesting.NewFilesystem(
			t,
			fstesting.WithDirectory(
				"default",
				fstesting.WithFile("features.yaml", segmentContents),
			),
		)

		segment := &core.Segment{
			Key:         "segment1",
			Name:        "Updated Segment",
			Description: "An updated test segment",
			MatchType:   core.MatchType_ANY_MATCH_TYPE,
		}

		any, err := newAny(segment)
		require.NoError(t, err)

		err = storage.PutResource(ctx, fs, &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          segment.Key,
			Payload:      any,
		})
		require.NoError(t, err)

		// verify segment was updated
		resource, err := storage.GetResource(ctx, fs, "default", segment.Key)
		require.NoError(t, err)

		var updated core.Segment
		err = resource.Payload.UnmarshalTo(&updated)
		require.NoError(t, err)
		assert.Equal(t, segment.Name, updated.Name)
		assert.Equal(t, segment.Description, updated.Description)
		assert.Equal(t, segment.MatchType, updated.MatchType)
	})
}

func TestSegmentStorage_DeleteResource(t *testing.T) {
	ctx := context.TODO()
	logger := zaptest.NewLogger(t)
	storage := NewSegmentStorage(logger)

	tests := []struct {
		name      string
		namespace string
		key       string
		wantErr   bool
	}{
		{
			name:      "delete existing segment",
			namespace: "default",
			key:       "segment1",
		},
		{
			name:      "delete non-existent segment",
			namespace: "default",
			key:       "nonexistent",
		},
		{
			name:      "delete from non-existent namespace",
			namespace: "nonexistent",
			key:       "segment1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := fstesting.NewFilesystem(
				t,
				fstesting.WithDirectory(
					"default",
					fstesting.WithFile("features.yaml", segmentContents),
				),
			)

			err := storage.DeleteResource(ctx, fs, tt.namespace, tt.key)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// verify segment was deleted
			_, err = storage.GetResource(ctx, fs, tt.namespace, tt.key)
			require.Error(t, err)
		})
	}
}

func TestSegmentStorage_DeleteResource_InUse(t *testing.T) {
	logger := zaptest.NewLogger(t)
	storage := NewSegmentStorage(logger)

	segmentInUseByRule := `version: "1.5"
namespace:
  key: default
  name: Default
segments:
  - key: segment1
    name: Segment 1
    match_type: ALL
flags:
  - key: flag1
    name: Flag 1
    type: VARIANT
    enabled: true
    variants:
      - key: var1
        name: Variant 1
    rules:
      - segment:
          keys:
            - segment1
          operator: OR
        distributions:
          - variant: var1
            rollout: 100
`

	segmentInUseByRollout := `version: "1.5"
namespace:
  key: default
  name: Default
segments:
  - key: segment1
    name: Segment 1
    match_type: ALL
flags:
  - key: flag1
    name: Flag 1
    type: BOOLEAN
    enabled: true
    rollouts:
      - description: Test rollout
        segment:
          keys:
            - segment1
          operator: OR
          value: true
`

	segmentInUseByMultipleFlags := `version: "1.5"
namespace:
  key: default
  name: Default
segments:
  - key: segment1
    name: Segment 1
    match_type: ALL
flags:
  - key: flag1
    name: Flag 1
    type: VARIANT
    enabled: true
    variants:
      - key: var1
        name: Variant 1
    rules:
      - segment:
          keys:
            - segment1
          operator: OR
        distributions:
          - variant: var1
            rollout: 100
  - key: flag2
    name: Flag 2
    type: BOOLEAN
    enabled: true
    rollouts:
      - description: Test rollout
        segment:
          keys:
            - segment1
          operator: OR
          value: true
`

	tests := []struct {
		name        string
		contents    string
		segmentKey  string
		wantErr     bool
		errContains string
	}{
		{
			name:        "cannot delete segment in use by rule",
			contents:    segmentInUseByRule,
			segmentKey:  "segment1",
			wantErr:     true,
			errContains: "segment \"segment1\" is in use by flag \"flag1\" rule 1",
		},
		{
			name:        "cannot delete segment in use by rollout",
			contents:    segmentInUseByRollout,
			segmentKey:  "segment1",
			wantErr:     true,
			errContains: "segment \"segment1\" is in use by flag \"flag1\" rollout 1",
		},
		{
			name:        "cannot delete segment in use by multiple flags",
			contents:    segmentInUseByMultipleFlags,
			segmentKey:  "segment1",
			wantErr:     true,
			errContains: "segment \"segment1\" is in use by flag \"flag1\" rule 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := fstesting.NewFilesystem(
				t,
				fstesting.WithDirectory(
					"default",
					fstesting.WithFile("features.yaml", tt.contents),
				),
			)

			err := storage.DeleteResource(t.Context(), fs, "default", tt.segmentKey)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
		})
	}
}
