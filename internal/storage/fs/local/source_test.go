package local

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap"
)

func Test_SourceString(t *testing.T) {
	assert.Equal(t, "local", (&Source{}).String())
}

func Test_SourceGet(t *testing.T) {
	s, err := NewSource(zap.NewNop(), "testdata", WithPollInterval(5*time.Second))
	assert.NoError(t, err)

	snap, err := s.Get(context.Background())
	assert.NoError(t, err)

	_, err = snap.GetNamespace(context.TODO(), "production")
	require.NoError(t, err)
}

func Test_SourceSubscribe(t *testing.T) {
	s, err := NewSource(zap.NewNop(), "testdata", WithPollInterval(5*time.Second))
	assert.NoError(t, err)

	dir, err := os.Getwd()
	assert.NoError(t, err)

	ftc := filepath.Join(dir, "testdata", "a.features.yml")

	defer func() {
		_, err := os.Stat(ftc)
		if err == nil {
			err := os.Remove(ftc)
			assert.NoError(t, err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan *storagefs.StoreSnapshot)
	go s.Subscribe(ctx, ch)

	// change the filesystem contents
	assert.NoError(t, os.WriteFile(ftc, []byte(`{"namespace":"staging"}`), os.ModePerm))

	select {
	case snap := <-ch:
		_, err := snap.GetNamespace(ctx, "staging")
		assert.NoError(t, err)
		cancel()

		_, open := <-ch
		assert.False(t, open, "expected channel to be closed after cancel")
	case <-time.After(10 * time.Second):
		t.Fatal("event not caught")
	}
}
