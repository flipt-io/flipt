package flipt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage/environments/fs"
	fstesting "go.flipt.io/flipt/internal/storage/environments/fs/testing"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	defaultContents = `version: "1.5"
namespace:
  key: default
  name: Default
  description: The default namespace
`

	teamAContents = `version: "1.5"
namespace:
  key: team_a
  name: Team A
  description: The Team A namespace
`

	multiDocContents = `version: "1.5"
namespace:
  key: team_b
  name: Team B
  description: The Team B namespace
---
version: "1.5"
namespace:
  key: team_c
  name: Team C
  description: The Team C namespace
`

	filesystem = func(t *testing.T) fs.Filesystem {
		return fstesting.NewFilesystem(
			t,
			fstesting.WithDirectory(
				"default",
				fstesting.WithFile("features.yaml", defaultContents),
			),
			fstesting.WithDirectory(
				"team_a",
				fstesting.WithFile("features.yaml", teamAContents),
			),
			fstesting.WithDirectory(
				"team_b",
				fstesting.WithFile("features.yaml", multiDocContents),
			),
		)
	}
)

func Test_getDocsAndNamespace(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name         string
		namespace    string
		wantDocs     int
		wantIndex    int
		wantErr      bool
		wantErrMatch string
	}{
		{
			name:      "default namespace",
			namespace: "default",
			wantDocs:  1,
			wantIndex: 0,
		},
		{
			name:      "team_a namespace",
			namespace: "team_a",
			wantDocs:  1,
			wantIndex: 0,
		},
		{
			name:      "multi-doc namespace",
			namespace: "team_b",
			wantDocs:  2,
			wantIndex: 0,
		},
		{
			name:         "non-existent namespace",
			namespace:    "nonexistent",
			wantErr:      true,
			wantErrMatch: "namespace not found: \"nonexistent\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, idx, err := getDocsAndNamespace(ctx, filesystem(t), tt.namespace)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMatch)
				return
			}

			require.NoError(t, err)
			assert.Len(t, docs, tt.wantDocs)
			assert.Equal(t, tt.wantIndex, idx)
			assert.Equal(t, tt.namespace, docs[idx].Namespace.GetKey())
		})
	}

	t.Run("auto-creates default namespace if not found", func(t *testing.T) {
		fs := fstesting.NewFilesystem(t) // empty filesystem
		docs, idx, err := getDocsAndNamespace(ctx, fs, "default")
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, 0, idx)
		assert.Equal(t, ext.DefaultNamespace, docs[0].Namespace)
	})
}

func Test_parseNamespace(t *testing.T) {
	ctx := context.TODO()

	tests := []struct {
		name         string
		namespace    string
		wantDocs     int
		wantErr      bool
		wantErrMatch string
	}{
		{
			name:      "default namespace",
			namespace: "default",
			wantDocs:  1,
		},
		{
			name:      "team_a namespace",
			namespace: "team_a",
			wantDocs:  1,
		},
		{
			name:      "multi-doc namespace",
			namespace: "team_b",
			wantDocs:  2,
		},
		{
			name:      "non-existent namespace returns empty docs",
			namespace: "nonexistent",
			wantDocs:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docs, err := parseNamespace(ctx, filesystem(t), tt.namespace)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMatch)
				return
			}

			require.NoError(t, err)
			assert.Len(t, docs, tt.wantDocs)

			if tt.wantDocs > 0 {
				// verify namespace key is set
				assert.NotEmpty(t, docs[0].Namespace.GetKey())
			}
		})
	}

	t.Run("handles invalid yaml", func(t *testing.T) {
		fs := fstesting.NewFilesystem(
			t,
			fstesting.WithDirectory(
				"invalid",
				fstesting.WithFile("features.yaml", "invalid: yaml: content"),
			),
		)

		docs, err := parseNamespace(ctx, fs, "invalid")
		require.Error(t, err)
		assert.Nil(t, docs)
	})

	t.Run("sets default namespace if empty in document", func(t *testing.T) {
		fs := fstesting.NewFilesystem(
			t,
			fstesting.WithDirectory(
				"empty_ns",
				fstesting.WithFile("features.yaml", "version: \"1.5\"\nnamespace: {}\n"),
			),
		)

		docs, err := parseNamespace(ctx, fs, "empty_ns")
		require.NoError(t, err)
		assert.Len(t, docs, 1)
		assert.Equal(t, ext.DefaultNamespace, docs[0].Namespace)
	})
}

func Test_newAny(t *testing.T) {
	tests := []struct {
		name    string
		msg     proto.Message
		wantErr bool
	}{
		{
			name: "valid message",
			msg:  &anypb.Any{TypeUrl: "type.googleapis.com/google.protobuf.Any"},
		},
		{
			name:    "nil message",
			msg:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			any, err := newAny(tt.msg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, any)
			// verify type URL prefix is stripped
			assert.NotContains(t, any.TypeUrl, "type.googleapis.com/")
		})
	}
}
