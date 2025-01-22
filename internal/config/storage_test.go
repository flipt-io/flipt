package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageConfigInfo(t *testing.T) {
	tests := []struct {
		config   StorageConfig
		expected map[string]string
	}{
		{StorageConfig{Type: GitStorageType, Git: &StorageGitConfig{Repository: "repo1", Ref: "v1.0.0"}}, map[string]string{
			"ref": "v1.0.0", "repository": "repo1",
		}},
	}

	for _, tt := range tests {
		t.Run(string(tt.config.Type), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.config.Info())
		})
	}
}
