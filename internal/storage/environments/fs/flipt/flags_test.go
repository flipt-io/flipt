package flipt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	fstesting "go.flipt.io/flipt/internal/storage/environments/fs/testing"
	"go.flipt.io/flipt/internal/storage/graph"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	emptyNamespaceContents = `version: "1.5"
namespace:
  key: default
  name: Default
  description: The default namespace
flags: []
`

	flagContents = `version: "1.5"
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
`

	multiDocFlagContents = `version: "1.5"
namespace:
  key: team_a
  name: Team A
  description: The Team A namespace
flags:
  - key: flag1
    name: Flag 1
    type: BOOLEAN
    enabled: true
---
version: "1.5"
namespace:
  key: team_a
  name: Team A
  description: The Team A namespace
flags:
  - key: flag2
    name: Flag 2
    type: STRING
    enabled: false
`
)

func TestFlagStorage_GetResource(t *testing.T) {
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = graph.NewDependencyGraph()
		storage         = NewFlagStorage(logger, dependencyGraph)
	)

	fs := fstesting.NewFilesystem(
		t,
		fstesting.WithDirectory(
			"default",
			fstesting.WithFile("features.yaml", flagContents),
		),
		fstesting.WithDirectory(
			"team_a",
			fstesting.WithFile("features.yaml", multiDocFlagContents),
		),
	)

	tests := []struct {
		name      string
		namespace string
		key       string
		wantErr   bool
	}{
		{
			name:      "get existing flag",
			namespace: "default",
			key:       "flag1",
		},
		{
			name:      "get flag from multi-doc",
			namespace: "team_a",
			key:       "flag2",
		},
		{
			name:      "flag not found",
			namespace: "default",
			key:       "nonexistent",
			wantErr:   true,
		},
		{
			name:      "namespace not found",
			namespace: "nonexistent",
			key:       "flag1",
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

			var flag core.Flag
			err = resource.Payload.UnmarshalTo(&flag)
			require.NoError(t, err)
			assert.Equal(t, tt.key, flag.Key)
		})
	}
}

func TestFlagStorage_ListResources(t *testing.T) {
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = graph.NewDependencyGraph()
		storage         = NewFlagStorage(logger, dependencyGraph)
	)

	fs := fstesting.NewFilesystem(
		t,
		fstesting.WithDirectory(
			"default",
			fstesting.WithFile("features.yaml", flagContents),
		),
		fstesting.WithDirectory(
			"team_a",
			fstesting.WithFile("features.yaml", multiDocFlagContents),
		),
	)

	tests := []struct {
		name      string
		namespace string
		want      int
		wantErr   bool
	}{
		{
			name:      "list flags in default namespace",
			namespace: "default",
			want:      1,
		},
		{
			name:      "list flags in multi-doc namespace",
			namespace: "team_a",
			want:      2,
		},
		{
			name:      "list flags in non-existent namespace",
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
				var flag core.Flag
				err = resource.Payload.UnmarshalTo(&flag)
				require.NoError(t, err)
			}
		})
	}
}

func TestFlagStorage_PutResource(t *testing.T) {
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = graph.NewDependencyGraph()
		storage         = NewFlagStorage(logger, dependencyGraph)
	)

	t.Run("create new flag", func(t *testing.T) {
		fs := fstesting.NewFilesystem(
			t,
			fstesting.WithDirectory(
				"default",
				fstesting.WithFile("features.yaml", emptyNamespaceContents),
			),
		)

		metadata, err := structpb.NewStruct(map[string]interface{}{
			"team": "backend",
		})
		require.NoError(t, err)

		flag := &core.Flag{
			Key:         "new_flag",
			Name:        "New Flag",
			Type:        core.FlagType_BOOLEAN_FLAG_TYPE,
			Description: "A new test flag",
			Enabled:     true,
			Metadata:    metadata,
		}

		any, err := newAny(flag)
		require.NoError(t, err)

		err = storage.PutResource(ctx, fs, &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          flag.Key,
			Payload:      any,
		})
		require.NoError(t, err)

		// verify flag was created
		resource, err := storage.GetResource(ctx, fs, "default", flag.Key)
		require.NoError(t, err)
		assert.Equal(t, flag.Key, resource.Key)

		var created core.Flag
		err = resource.Payload.UnmarshalTo(&created)
		require.NoError(t, err)
		assert.Equal(t, flag.Name, created.Name)
		assert.Equal(t, flag.Description, created.Description)
		assert.Equal(t, flag.Enabled, created.Enabled)
		assert.Equal(t, flag.Type, created.Type)
		assert.Equal(t, flag.Metadata.AsMap(), created.Metadata.AsMap())
	})

	t.Run("update existing flag", func(t *testing.T) {
		fs := fstesting.NewFilesystem(
			t,
			fstesting.WithDirectory(
				"default",
				fstesting.WithFile("features.yaml", flagContents),
			),
		)

		flag := &core.Flag{
			Key:         "flag1",
			Name:        "Updated Flag",
			Type:        core.FlagType_BOOLEAN_FLAG_TYPE,
			Description: "An updated test flag",
			Enabled:     false,
		}

		any, err := newAny(flag)
		require.NoError(t, err)

		err = storage.PutResource(ctx, fs, &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          flag.Key,
			Payload:      any,
		})
		require.NoError(t, err)

		// verify flag was updated
		resource, err := storage.GetResource(ctx, fs, "default", flag.Key)
		require.NoError(t, err)

		var updated core.Flag
		err = resource.Payload.UnmarshalTo(&updated)
		require.NoError(t, err)
		assert.Equal(t, flag.Name, updated.Name)
		assert.Equal(t, flag.Description, updated.Description)
		assert.Equal(t, flag.Enabled, updated.Enabled)
		assert.Equal(t, flag.Type, updated.Type)
	})
}

func TestFlagStorage_DeleteResource(t *testing.T) {
	var (
		ctx             = context.TODO()
		logger          = zaptest.NewLogger(t)
		dependencyGraph = graph.NewDependencyGraph()
		storage         = NewFlagStorage(logger, dependencyGraph)
	)

	tests := []struct {
		name      string
		namespace string
		key       string
		wantErr   bool
	}{
		{
			name:      "delete existing flag",
			namespace: "default",
			key:       "flag1",
		},
		{
			name:      "delete non-existent flag",
			namespace: "default",
			key:       "nonexistent",
		},
		{
			name:      "delete from non-existent namespace",
			namespace: "nonexistent",
			key:       "flag1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := fstesting.NewFilesystem(
				t,
				fstesting.WithDirectory(
					"default",
					fstesting.WithFile("features.yaml", flagContents),
				),
			)

			err := storage.DeleteResource(ctx, fs, tt.namespace, tt.key)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// verify flag was deleted
			_, err = storage.GetResource(ctx, fs, tt.namespace, tt.key)
			require.Error(t, err)
		})
	}
}
