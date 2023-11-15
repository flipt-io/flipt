//go:build !linux
// +build !linux

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDatabaseRoot(t *testing.T) {
	root, err := defaultDatabaseRoot()
	require.NoError(t, err)

	configDir, err := os.UserConfigDir()
	require.NoError(t, err)

	assert.Equal(t, root, filepath.Join(configDir, "flipt"))
}
