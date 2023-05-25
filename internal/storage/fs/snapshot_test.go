package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/internal/storage/fs/fixtures/fswithindex"
	"go.flipt.io/flipt/internal/storage/fs/fixtures/fswithoutindex"
	"go.uber.org/zap"
)

func TestFSWithIndex(t *testing.T) {
	filenames, exclusions, err := buildSnapshotHelper(zap.NewNop(), fswithindex.FS)
	assert.NoError(t, err)

	assert.Len(t, exclusions, 1)
	assert.Len(t, filenames, 3)
}

func TestFSWithoutIndex(t *testing.T) {
	filenames, exclusions, err := buildSnapshotHelper(zap.NewNop(), fswithoutindex.FS)
	assert.NoError(t, err)

	assert.Len(t, exclusions, 0)
	assert.Len(t, filenames, 6)
}
