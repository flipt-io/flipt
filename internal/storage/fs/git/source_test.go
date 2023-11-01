package git

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/containers"
	storagefs "go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap/zaptest"
)

var gitRepoURL = os.Getenv("TEST_GIT_REPO_URL")

func Test_SourceString(t *testing.T) {
	require.Equal(t, "git", (&Source{}).String())
}

func Test_SourceGet(t *testing.T) {
	source, skip := testSource(t)
	if skip {
		return
	}

	snap, err := source.Get()
	require.NoError(t, err)

	_, err = snap.GetNamespace(context.TODO(), "production")
	require.NoError(t, err)
}

func Test_SourceSubscribe_Hash(t *testing.T) {
	head := os.Getenv("TEST_GIT_REPO_HEAD")
	if head == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_HEAD env var to run this test.")
		return
	}

	source, skip := testSource(t, WithRef(head))
	if skip {
		return
	}

	ch := make(chan *storagefs.StoreSnapshot)
	source.Subscribe(context.Background(), ch)

	_, closed := <-ch
	assert.False(t, closed, "expected channel to be closed")
}

func Test_SourceSubscribe(t *testing.T) {
	source, skip := testSource(t)
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	// prime source
	_, err := source.Get()
	require.NoError(t, err)

	// start subscription
	ch := make(chan *storagefs.StoreSnapshot)
	go source.Subscribe(ctx, ch)

	// pull repo
	workdir := memfs.New()
	repo, err := git.Clone(memory.NewStorage(), workdir, &git.CloneOptions{
		Auth:          &http.BasicAuth{Username: "root", Password: "password"},
		URL:           gitRepoURL,
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName("main"),
	})
	require.NoError(t, err)

	tree, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, tree.Checkout(&git.CheckoutOptions{
		Branch: "refs/heads/main",
	}))

	// update features.yml
	fi, err := workdir.OpenFile("features.yml", os.O_TRUNC|os.O_RDWR, os.ModePerm)
	require.NoError(t, err)

	updated := []byte(`namespace: production
flags:
    - key: foo
      name: Foo`)

	_, err = fi.Write(updated)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	// commit changes
	_, err = tree.Commit("chore: update features.yml", &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	// push new commit
	require.NoError(t, repo.Push(&git.PushOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
	}))

	// assert matching state
	var snap *storagefs.StoreSnapshot
	select {
	case snap = <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for snapshot")
	}

	require.NoError(t, err)

	t.Log("received new snapshot")

	_, err = snap.GetFlag(ctx, "production", "foo")
	require.NoError(t, err)

	// ensure closed
	cancel()

	_, open := <-ch
	require.False(t, open, "expected channel to be closed after cancel")
}

func testSource(t *testing.T, opts ...containers.Option[Source]) (*Source, bool) {
	t.Helper()

	if gitRepoURL == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_URL env var to run this test.")
		return nil, true
	}

	source, err := NewSource(zaptest.NewLogger(t), gitRepoURL,
		append([]containers.Option[Source]{
			WithRef("main"),
			WithPollInterval(5 * time.Second),
			WithAuth(&http.BasicAuth{
				Username: "root",
				Password: "password",
			}),
		},
			opts...)...,
	)
	require.NoError(t, err)

	return source, false
}
