package git

import (
	"flag"
	"io"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var gitRepoURL = flag.String("git-repo-url", "", "Target repository to use for testing")

func Test_SourceGet(t *testing.T) {
	if *gitRepoURL == "" {
		t.Skip("Set non-empty --git-repo-url to run this test.")
		return
	}

	source, err := NewSource(zaptest.NewLogger(t), *gitRepoURL,
		WithRef("main"),
		WithPollInterval(5*time.Second),
		WithAuth(&http.BasicAuth{
			Username: "root",
			Password: "password",
		}),
	)
	require.NoError(t, err)

	fs, err := source.Get()
	require.NoError(t, err)

	fi, err := fs.Open("features.yml")
	require.NoError(t, err)

	data, err := io.ReadAll(fi)
	require.NoError(t, err)

	assert.Equal(t, []byte("namespace: production\n"), data)
}