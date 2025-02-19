package fs_test

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage/environments/fs"
	fstesting "go.flipt.io/flipt/internal/storage/environments/fs/testing"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap/zaptest"
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
		)
	}
)

func Test_NamespaceStorage_GetNamespace(t *testing.T) {
	ctx := context.TODO()
	logger := zaptest.NewLogger(t)
	storage := fs.NewNamespaceStorage(logger)

	t.Run("default namespace", func(t *testing.T) {
		ns, err := storage.GetNamespace(ctx, filesystem(t), "default")
		require.NoError(t, err)

		assert.Equal(t, &rpcenvironments.Namespace{
			Key:         "default",
			Name:        "Default",
			Description: ptr("The default namespace"),
			Protected:   ptr(true),
		}, ns)
	})

	t.Run("team_a namespace", func(t *testing.T) {
		ns, err := storage.GetNamespace(ctx, filesystem(t), "team_a")
		require.NoError(t, err)

		assert.Equal(t, &rpcenvironments.Namespace{
			Key:         "team_a",
			Name:        "Team A",
			Description: ptr("The Team A namespace"),
			Protected:   ptr(false),
		}, ns)
	})

	t.Run("team_b namespace (does not exist)", func(t *testing.T) {
		_, err := storage.GetNamespace(ctx, filesystem(t), "team_b")
		var notfound errors.ErrNotFound
		require.ErrorAs(t, err, &notfound)
	})
}

func Test_NamespaceStorage_ListNamespaces(t *testing.T) {
	ctx := context.TODO()
	logger := zaptest.NewLogger(t)
	storage := fs.NewNamespaceStorage(logger)

	items, err := storage.ListNamespaces(ctx, filesystem(t))
	require.NoError(t, err)

	assert.Equal(t, []*rpcenvironments.Namespace{
		{
			Key:         "default",
			Name:        "Default",
			Description: ptr("The default namespace"),
			Protected:   ptr(true),
		},
		{
			Key:         "team_a",
			Name:        "Team A",
			Description: ptr("The Team A namespace"),
			Protected:   ptr(false),
		},
	}, items)

	t.Run("empty sub-directory returns empty namespace list", func(t *testing.T) {
		items, err := storage.ListNamespaces(ctx, fs.SubFilesystem(filesystem(t), "notfound"))
		require.NoError(t, err)

		assert.Equal(t, []*rpcenvironments.Namespace{}, items)
	})
}

func Test_NamespaceStorage_PutNamespace(t *testing.T) {
	ctx := context.TODO()
	logger := zaptest.NewLogger(t)
	storage := fs.NewNamespaceStorage(logger)

	fs := filesystem(t)
	require.NoError(t, storage.PutNamespace(ctx, fs, &rpcenvironments.Namespace{
		Key:         "team_b",
		Name:        "Team B",
		Description: ptr("The Team B namespace"),
	}))

	t.Run("ensure file has expected contents", func(t *testing.T) {
		fi, err := fs.OpenFile("team_b/features.yaml", os.O_RDONLY, 0644)
		require.NoError(t, err)

		data, err := io.ReadAll(fi)
		require.NoError(t, err)

		assert.Equal(t, `version: "1.5"
namespace:
  key: team_b
  name: Team B
  description: The Team B namespace
`, string(data))

		require.NoError(t, fi.Close())
	})

	t.Run("ensure namespace is now retrievable", func(t *testing.T) {
		ns, err := storage.GetNamespace(ctx, fs, "team_b")
		require.NoError(t, err)

		assert.Equal(t, "team_b", ns.Key)
	})
}

func Test_NamespaceStorage_DeleteNamespace(t *testing.T) {
	ctx := context.TODO()
	logger := zaptest.NewLogger(t)
	storage := fs.NewNamespaceStorage(logger)

	fs := filesystem(t)
	require.NoError(t, storage.DeleteNamespace(ctx, fs, "team_a"))

	t.Run("ensure namespace is not found", func(t *testing.T) {
		_, err := storage.GetNamespace(ctx, fs, "team_a")
		var notfound errors.ErrNotFound
		require.ErrorAs(t, err, &notfound)
	})
}

func ptr[T any](t T) *T {
	return &t
}
