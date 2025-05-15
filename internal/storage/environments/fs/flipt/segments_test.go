package flipt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
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

	dependentSegmentContents = `version: "1.5"
namespace:
	key: default
	name: Default
	description: The default namespace
flags:
	- key: flag1
	name: Flag 1
	type: BOOLEAN
	description: A test flag
	enabled: true
	metadata:
		team: backend
	variants:
		- key: variant1
		name: Variant 1
		description: A test variant
		default: true
		attachment:
			color: blue
	rules:
		- segment:
			keys: [segment1]
			operator: AND
		distributions:
			- rollout: 100
			variant_key: variant1
	rollouts:
		- description: A test rollout
		segment:
			keys: [segment1]
			operator: AND
			value: true
segments:
	- key: segment1
	name: Segment 1
	match_type: ALL
`
)

func TestSegmentStorage_GetResource(t *testing.T) {
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = NewDependencyGraph()
		storage         = NewSegmentStorage(logger, dependencyGraph)
	)

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
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = NewDependencyGraph()
		storage         = NewSegmentStorage(logger, dependencyGraph)
	)

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
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = NewDependencyGraph()
		storage         = NewSegmentStorage(logger, dependencyGraph)
	)

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
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = NewDependencyGraph()
		storage         = NewSegmentStorage(logger, dependencyGraph)
	)

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

func TestSegmentStorage_DeleteResource_Dependent(t *testing.T) {
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = NewDependencyGraph()

		storage = NewSegmentStorage(logger, dependencyGraph)
	)

	// manually add a dependency between a flag and a segment for the test
	dependencyGraph.AddDependency(ResourceID{
		Namespace: "default",
		Key:       "flag1",
		Type:      serverenvironments.FlagResourceType,
	}, ResourceID{
		Namespace: "default",
		Key:       "segment1",
		Type:      serverenvironments.SegmentResourceType,
	})

	fs := fstesting.NewFilesystem(
		t,
		fstesting.WithDirectory(
			"default",
			fstesting.WithFile("features.yaml", dependentSegmentContents),
		),
	)

	err := storage.DeleteResource(ctx, fs, "default", "segment1")
	require.Error(t, err)
	assert.EqualError(t, err, "deleting segment default/segment1: segment cannot be deleted as it is a dependency of default/flipt.core.Flag/flag1")
}
