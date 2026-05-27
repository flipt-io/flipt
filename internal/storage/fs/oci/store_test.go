package oci

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/containers"
	fliptoci "go.flipt.io/flipt/internal/oci"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap/zaptest"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
)

func Test_SourceString(t *testing.T) {
	require.Equal(t, "oci", (&SnapshotStore{}).String())
}

func Test_SourceSubscribe(t *testing.T) {
	ch := make(chan struct{})
	store, target := testStore(t, WithPollOptions(
		fs.WithInterval(time.Second),
		fs.WithNotify(t, func(modified bool) {
			if modified {
				close(ch)
			}
		}),
	))

	ctx := context.Background()

	require.NoError(t, store.View(ctx, func(s storage.ReadOnlyStore) error {
		_, err := s.GetNamespace(ctx, storage.NewNamespace("production"))
		require.NoError(t, err)

		_, err = s.GetFlag(ctx, storage.NewResource("production", "foo"))
		require.Error(t, err, "should error as flag should not exist yet")

		return nil
	}))

	updateRepoContents(t, target,
		layer(
			"production",
			`{"namespace":"production","flags":[{"key":"foo","name":"Foo"}]}`,
			fliptoci.MediaTypeFliptNamespace,
		),
	)

	t.Log("waiting for new snapshot")

	// assert matching state
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for snapshot")
	}

	t.Log("received new snapshot")

	require.NoError(t, store.View(ctx, func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "foo"))
		require.NoError(t, err)
		return nil
	}))
}

func Test_Store_ContextWithSnapshot(t *testing.T) {
	logger := zaptest.NewLogger(t)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(dir+"/features.yml", []byte(`namespace: testing
flags:
    - key: test-flag
      name: Test Flag
`), 0600))

	snap, err := fs.SnapshotFromFS(logger, os.DirFS(dir))
	require.NoError(t, err)
	require.NotNil(t, snap)

	store := &SnapshotStore{
		snap: snap,
	}

	t.Run("captures snapshot in context", func(t *testing.T) {
		pinnedCtx := store.ContextWithSnapshot(t.Context())

		err := store.View(pinnedCtx, func(s storage.ReadOnlyStore) error {
			flag, err := s.GetFlag(t.Context(), storage.NewResource("testing", "test-flag"))
			require.NoError(t, err)
			require.Equal(t, "Test Flag", flag.Name)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("pinned snapshot in context is valid ReadOnlyStore", func(t *testing.T) {
		pinnedCtx := store.ContextWithSnapshot(t.Context())

		snap, ok := pinnedCtx.Value(snapshotCtxKey).(storage.ReadOnlyStore)
		require.True(t, ok)
		require.NotNil(t, snap)
	})

	t.Run("no pinned snapshot without ContextWithSnapshot", func(t *testing.T) {
		_, ok := t.Context().Value(snapshotCtxKey).(storage.ReadOnlyStore)
		require.False(t, ok)
	})
}

func testStore(t *testing.T, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, oras.Target) {
	t.Helper()

	target, dir, repo := testRepository(t,
		layer("production", `{"namespace":"production"}`, fliptoci.MediaTypeFliptNamespace),
	)

	store, err := fliptoci.NewStore(zaptest.NewLogger(t), dir)
	require.NoError(t, err)

	ref, err := fliptoci.ParseReference(fmt.Sprintf("flipt://local/%s:latest", repo))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	source, err := NewSnapshotStore(ctx,
		zaptest.NewLogger(t),
		store,
		ref,
		opts...)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = source.Close()
	})

	return source, target
}

func layer(ns, payload, mediaType string) func(*testing.T, oras.Target) v1.Descriptor {
	return func(t *testing.T, store oras.Target) v1.Descriptor {
		t.Helper()

		desc := v1.Descriptor{
			Digest:    digest.FromString(payload),
			Size:      int64(len(payload)),
			MediaType: mediaType,
			Annotations: map[string]string{
				fliptoci.AnnotationFliptNamespace: ns,
			},
		}

		require.NoError(t, store.Push(context.TODO(), desc, bytes.NewReader([]byte(payload))))

		return desc
	}
}

func testRepository(t *testing.T, layerFuncs ...func(*testing.T, oras.Target) v1.Descriptor) (oras.Target, string, string) {
	t.Helper()

	var (
		repository = "testrepo"
		dir        = t.TempDir()
	)

	store, err := oci.New(path.Join(dir, repository))
	require.NoError(t, err)

	store.AutoSaveIndex = true

	updateRepoContents(t, store, layerFuncs...)

	return store, dir, repository
}

func updateRepoContents(t *testing.T, target oras.Target, layerFuncs ...func(*testing.T, oras.Target) v1.Descriptor) {
	t.Helper()
	ctx := context.TODO()

	var layers []v1.Descriptor
	for _, fn := range layerFuncs {
		layers = append(layers, fn(t, target))
	}

	desc, err := oras.PackManifest(ctx, target, oras.PackManifestVersion1_1_RC4, fliptoci.MediaTypeFliptFeatures, oras.PackManifestOptions{
		ManifestAnnotations: map[string]string{},
		Layers:              layers,
	})
	require.NoError(t, err)

	require.NoError(t, target.Tag(ctx, desc, "latest"))
}
