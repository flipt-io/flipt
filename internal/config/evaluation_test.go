package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluationConfig_LoadFromYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	yaml := `version: "2.0"
evaluation:
  include_flag_metadata: true
`
	require.NoError(t, os.WriteFile(configPath, []byte(yaml), 0o600))

	result, err := Load(t.Context(), configPath)
	require.NoError(t, err)

	assert.True(t, result.Config.Evaluation.IncludeFlagMetadata)
}
