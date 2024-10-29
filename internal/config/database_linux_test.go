//go:build linux
// +build linux

package config

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDatabaseRoot(t *testing.T) {
	cfgDir, err := Dir()
	require.NoError(t, err)

	root, err := defaultDatabaseRoot()
	require.NoError(t, err)
	assert.Equal(t, cfgDir, root)
}

func TestFindDatabaseRoot(t *testing.T) {
	mockFS := afero.NewMemMapFs()
	err := mockFS.MkdirAll("/var/opt/flipt", 0000)
	require.NoError(t, err)

	root, err := findDatabaseRoot(mockFS)
	require.NoError(t, err)
	assert.Equal(t, "/var/opt/flipt", root)
}
