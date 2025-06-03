package environments

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
			owner, repo, err := parseGiteaRepo(ex)
			require.NoError(t, err)
			assert.Equal(t, "admin", owner)
			assert.Equal(t, "demo-repo", repo)
		})
	}
}
