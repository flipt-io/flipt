package local

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func Test_Store_String(t *testing.T) {
	assert.Equal(t, "local", (&SnapshotStore{}).String())
}

func Test_Store(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var closed bool
	ch := make(chan struct{})

	s, err := NewSnapshotStore(ctx, zap.NewNop(), "testdata", WithPollOptions(
		storagefs.WithInterval(1*time.Second),
		storagefs.WithNotify(t, func(modified bool) {
			if modified && !closed {
				closed = true
				close(ch)
			}
		}),
	))
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = s.Close()
	})

	dir, err := os.Getwd()
	require.NoError(t, err)

	ftc := filepath.Join(dir, "testdata", "a.features.yml")

	defer func() {
		_, err := os.Stat(ftc)
		if err == nil {
			err := os.Remove(ftc)
			assert.NoError(t, err)
		}
	}()

	// change the filesystem contents
	assert.NoError(t, os.WriteFile(ftc, []byte(`{"namespace":"staging"}`), 0600))

	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		t.Fatal("event not caught")
	}

	assert.NoError(t, s.View(ctx, func(s storage.ReadOnlyStore) error {
		_, err = s.GetNamespace(ctx, storage.NewNamespace("staging"))
		return err
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

	snap, err := storagefs.SnapshotFromFS(logger, os.DirFS(dir))
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
