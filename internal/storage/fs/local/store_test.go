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
