package local

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/storage"
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
		WithInterval(1*time.Second),
		WithNotify(t, func() {
			if !closed {
				closed = true
				close(ch)
			}
		}),
	))
	assert.NoError(t, err)

	t.Cleanup(func() {
		_ = s.Close()
	})

	dir, err := os.Getwd()
	assert.NoError(t, err)

	ftc := filepath.Join(dir, "testdata", "staging", "features.yaml")

	defer func() {
		_, err := os.Stat(ftc)
		if err == nil {
			err := os.Remove(ftc)
			assert.NoError(t, err)
		}
	}()

	// change the filesystem contents
	assert.NoError(t, os.WriteFile(ftc, []byte(`{"namespace":"staging"}`), os.ModePerm))

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
