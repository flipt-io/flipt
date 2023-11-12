//go:build linux
// +build linux

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultDatabaseRoot(t *testing.T) {
	root, err := defaultDatabaseRoot()
	require.NoError(t, err)
	assert.Equal(t, "/var/opt/flipt", root)
}
