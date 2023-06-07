package local

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_SourceGet(t *testing.T) {
	s, err := NewSource(zap.NewNop(), "testdata")
	assert.NoError(t, err)

	tfs, err := s.Get()
	assert.NoError(t, err)

	file, err := tfs.Open("features.yml")
	assert.NoError(t, err)

	assert.NotNil(t, file)
}

func Test_SourceSubscribe(t *testing.T) {
	s, err := NewSource(zap.NewNop(), "testdata")
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

	fsCh := make(chan fs.FS)
	go s.Subscribe(context.Background(), fsCh)

	// Create event
	_, err = os.Create(ftc)
	assert.NoError(t, err)

	select {
	case f := <-fsCh:
		file, err := f.Open("features.yml")
		assert.NoError(t, err)
		assert.NotNil(t, file)
	case <-time.After(10 * time.Second):
		t.Fatal("event not caught")
	}
}
