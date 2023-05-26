package fs

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

//go:embed all:fixtures
var testdata embed.FS

func TestFSWithIndex(t *testing.T) {
	fwi, _ := fs.Sub(testdata, "fixtures/fswithindex")

	filenames, err := buildSnapshotHelper(zap.NewNop(), fwi)
	assert.NoError(t, err)

	expected := []string{
		"prod/prod.features.yml",
		"sandbox/sandbox.features.yaml",
	}
	assert.Len(t, filenames, 2)
	assert.ElementsMatch(t, filenames, expected)
}

func TestFSWithoutIndex(t *testing.T) {
	fwoi, _ := fs.Sub(testdata, "fixtures/fswithoutindex")
	filenames, err := buildSnapshotHelper(zap.NewNop(), fwoi)
	assert.NoError(t, err)

	expected := []string{
		"prod/prod.features.yaml",
		"prod/features.yml",
		"sandbox/sandbox.features.yml",
		"sandbox/features.yaml",
		"staging/staging.features.yaml",
		"staging/features.yml",
	}
	assert.Len(t, filenames, 6)
	assert.ElementsMatch(t, filenames, expected)
}
