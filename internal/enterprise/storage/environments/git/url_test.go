package git

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGiteaRepo(t *testing.T) {
	tests := []string{
		"https://gitea.example.com/admin/demo-repo.git",
		"git@gitea.example.com:admin/demo-repo.git",
		"git://gitea.example.com/admin/demo-repo.git",
	}

	for i, ex := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			gitURL, err := ParseGitURL(ex)
			require.NoError(t, err)
			assert.Equal(t, "admin", gitURL.Owner)
			assert.Equal(t, "demo-repo", gitURL.Repo)
		})
	}
}

func TestParseGithubRepo(t *testing.T) {
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
}
