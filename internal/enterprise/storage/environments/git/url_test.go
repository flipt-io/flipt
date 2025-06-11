package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGitURL(t *testing.T) {
	t.Run("gitea", func(t *testing.T) {
		tests := []string{
			"https://gitea.example.com/flipt-io/flipt.git",
			"git@gitea.example.com:flipt-io/flipt.git",
			"git://gitea.example.com/flipt-io/flipt.git",
		}

		for i, ex := range tests {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				gitURL, err := ParseGitURL(ex)
				require.NoError(t, err)
				assert.Equal(t, "flipt-io", gitURL.Owner)
				assert.Equal(t, "flipt", gitURL.Repo)
			})
		}
	})

	t.Run("github", func(t *testing.T) {
		tests := []string{
			"https://github.com/flipt-io/flipt",
			"https://www.github.com/flipt-io/flipt",
			"git@github.com:flipt-io/flipt.git",
		}

		for i, ex := range tests {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				gitURL, err := ParseGitURL(ex)
				require.NoError(t, err)
				assert.Equal(t, "flipt-io", gitURL.Owner)
				assert.Equal(t, "flipt", gitURL.Repo)
			})
		}
	})

	t.Run("gitlab", func(t *testing.T) {
		tests := []string{
			"https://gitlab.com/flipt-io/flipt",
			"https://www.gitlab.com/flipt-io/flipt",
			"git@gitlab.com:flipt-io/flipt.git",
		}

		for i, ex := range tests {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				gitURL, err := ParseGitURL(ex)
				require.NoError(t, err)
				assert.Equal(t, "flipt-io", gitURL.Owner)
				assert.Equal(t, "flipt", gitURL.Repo)
			})
		}
	})
}
