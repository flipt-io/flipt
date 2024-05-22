package git

import (
	"bytes"
	"context"
	"encoding/pem"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/fs"
	"go.uber.org/zap/zaptest"
)

var gitRepoURL = os.Getenv("TEST_GIT_REPO_URL")

func Test_Store_String(t *testing.T) {
	require.Equal(t, "git", (&SnapshotStore{}).String())
}

func Test_Store_Subscribe_Hash(t *testing.T) {
	head := os.Getenv("TEST_GIT_REPO_HEAD")
	if head == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_HEAD env var to run this test.")
		return
	}

	// this helper will fail if there is a problem with this option
	// the only difference in behaviour is that the poll loop
	// will silently (intentionally) not run
	testStore(t, gitRepoURL, WithRef(head))
}

func Test_Store_View(t *testing.T) {
	ch := make(chan struct{})
	store, skip := testStore(t, gitRepoURL, WithPollOptions(
		fs.WithInterval(time.Second),
		fs.WithNotify(t, func(modified bool) {
			if modified {
				close(ch)
			}
		}),
	))
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

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

	// wait until the snapshot is updated or
	// we timeout
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for snapshot")
	}

	require.NoError(t, err)

	t.Log("received new snapshot")

	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err = s.GetFlag(ctx, storage.NewResource("production", "foo"))
		return err
	}))
}

func Test_Store_View_WithFilesystemStorage(t *testing.T) {
	ch := make(chan struct{})
	store, skip := testStore(t, gitRepoURL,
		WithFilesystemStorage(t.TempDir()),
		WithPollOptions(
			fs.WithInterval(time.Second),
			fs.WithNotify(t, func(modified bool) {
				if modified {
					close(ch)
				}
			}),
		))
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

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

	// wait until the snapshot is updated or
	// we timeout
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for snapshot")
	}

	require.NoError(t, err)

	t.Log("received new snapshot")

	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err = s.GetFlag(ctx, storage.NewResource("production", "foo"))
		return err
	}))
}

func Test_Store_View_WithRevision(t *testing.T) {
	ch := make(chan struct{})
	store, skip := testStore(t, gitRepoURL, WithPollOptions(
		fs.WithInterval(time.Second),
		fs.WithNotify(t, func(modified bool) {
			if modified {
				close(ch)
			}
		}),
	))
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

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
		Branch: "refs/heads/new-branch",
		Create: true,
	}))

	// update features.yml
	fi, err := workdir.OpenFile("features.yml", os.O_TRUNC|os.O_RDWR, os.ModePerm)
	require.NoError(t, err)

	updated := []byte(`namespace: production
flags:
  - key: foo
    name: Foo
  - key: bar
    name: Bar`)

	_, err = fi.Write(updated)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	// commit changes
	_, err = tree.Commit("chore: update features.yml add foo and bar", &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	// push new commit
	require.NoError(t, repo.Push(&git.PushOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/new-branch:refs/heads/new-branch"},
	}))

	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "bar"))
		require.Error(t, err, "flag should not be found in default revision")
		return nil
	}))

	require.NoError(t, store.View(ctx, "main", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "bar"))
		require.Error(t, err, "flag should not be found in explicitly named main revision")
		return nil
	}))

	// should be able to fetch flag from previously unfetched reference
	require.NoError(t, store.View(ctx, "new-branch", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "bar"))
		require.NoError(t, err, "flag should be present on new-branch")
		return nil
	}))

	// flag bar should not yet be present
	require.NoError(t, store.View(ctx, "new-branch", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "baz"))
		require.Error(t, err, "flag should not be found in explicitly named new-branch revision")
		return nil
	}))

	// update features.yml, now with the bar flag
	fi, err = workdir.OpenFile("features.yml", os.O_TRUNC|os.O_RDWR, os.ModePerm)
	require.NoError(t, err)

	updated = []byte(`namespace: production
flags:
  - key: foo
    name: Foo
  - key: bar
    name: Bar
  - key: baz
    name: Baz`)

	_, err = fi.Write(updated)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	// commit changes
	_, err = tree.Commit("chore: update features.yml add baz", &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	// push new commit
	require.NoError(t, repo.Push(&git.PushOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/new-branch:refs/heads/new-branch"},
	}))

	// we should expect to see a modified event now because
	// the new reference should be tracked
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for fetch")
	}

	// should be able to fetch flag bar now that it has been pushed
	require.NoError(t, store.View(ctx, "new-branch", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "baz"))
		require.NoError(t, err, "flag should be present on new-branch")
		return nil
	}))
}

func Test_Store_View_WithSemverRevision(t *testing.T) {
	tag := os.Getenv("TEST_GIT_REPO_TAG")
	if tag == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_TAG env var to run this test.")
		return
	}

	head := os.Getenv("TEST_GIT_REPO_HEAD")
	if head == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_HEAD env var to run this test.")
		return
	}

	ch := make(chan struct{})
	store, skip := testStore(t, gitRepoURL,
		WithRef("v0.1.*"),
		WithSemverResolver(),
		WithPollOptions(
			fs.WithInterval(time.Second),
			fs.WithNotify(t, func(modified bool) {
				if modified {
					close(ch)
				}
			}),
		),
	)
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

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
		Branch: "refs/heads/semver-branch",
		Create: true,
	}))

	// update features.yml
	fi, err := workdir.OpenFile("features.yml", os.O_TRUNC|os.O_RDWR, os.ModePerm)
	require.NoError(t, err)

	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("semver", "bar"))
		require.Error(t, err, "flag should not be found in default revision")
		return nil
	}))

	updated := []byte(`namespace: semver
flags:
  - key: foo
    name: Foo
  - key: bar
    name: Bar`)

	_, err = fi.Write(updated)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	// commit changes
	commit, err := tree.Commit("chore: update features.yml", &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	// create a new tag respecting the semver constraint
	_, err = repo.CreateTag("v0.1.4", commit, nil)
	require.NoError(t, err)

	// push new commit
	require.NoError(t, repo.Push(&git.PushOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			"refs/heads/semver-branch:refs/heads/semver-branch",
			"refs/tags/v0.1.4:refs/tags/v0.1.4",
		},
	}))

	// wait until the snapshot is updated or
	// we timeout
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for snapshot")
	}

	require.NoError(t, err)

	t.Log("received new snapshot")

	// Test if we can resolve to the new tag
	hash, err := store.resolve("v0.1.*")
	require.NoError(t, err)
	require.Equal(t, commit.String(), hash.String())

	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("semver", "bar"))
		require.NoError(t, err, "flag should be present on semver v0.1.*")
		return nil
	}))
}

func Test_Store_View_WithDirectory(t *testing.T) {
	ch := make(chan struct{})
	store, skip := testStore(t, gitRepoURL, WithPollOptions(
		fs.WithInterval(time.Second),
		fs.WithNotify(t, func(modified bool) {
			if modified {
				close(ch)
			}
		}),
	),
		// scope flag state discovery to sub-directory
		WithDirectory("subdir"),
	)
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

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

	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err = s.GetFlag(ctx, storage.NewResource("alternative", "otherflag"))
		return err
	}))
}

func Test_Store_SelfSignedSkipTLS(t *testing.T) {
	ts := httptest.NewTLSServer(nil)
	defer ts.Close()
	// This is not a valid Git source, but it still proves the point that a
	// well-known server with a self-signed certificate will be accepted by Flipt
	// when configuring the TLS options for the source
	err := testStoreWithError(t, ts.URL, WithInsecureTLS(false))
	require.ErrorContains(t, err, "tls: failed to verify certificate: x509: certificate signed by unknown authority")
	err = testStoreWithError(t, ts.URL, WithInsecureTLS(true))
	// This time, we don't expect a tls validation error anymore
	require.ErrorIs(t, err, transport.ErrRepositoryNotFound)
}

func Test_Store_SelfSignedCABytes(t *testing.T) {
	ts := httptest.NewTLSServer(nil)
	defer ts.Close()
	var buf bytes.Buffer
	pemCert := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ts.Certificate().Raw,
	}
	err := pem.Encode(&buf, pemCert)
	require.NoError(t, err)

	// This is not a valid Git source, but it still proves the point that a
	// well-known server with a self-signed certificate will be accepted by Flipt
	// when configuring the TLS options for the source
	err = testStoreWithError(t, ts.URL)
	require.ErrorContains(t, err, "tls: failed to verify certificate: x509: certificate signed by unknown authority")
	err = testStoreWithError(t, ts.URL, WithCABundle(buf.Bytes()))
	// This time, we don't expect a tls validation error anymore
	require.ErrorIs(t, err, transport.ErrRepositoryNotFound)
}

func testStore(t *testing.T, gitRepoURL string, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, bool) {
	t.Helper()

	if gitRepoURL == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_URL env var to run this test.")
		return nil, true
	}

	t.Log("Git repo host:", gitRepoURL)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	source, err := NewSnapshotStore(ctx, zaptest.NewLogger(t), gitRepoURL,
		append([]containers.Option[SnapshotStore]{
			WithRef("main"),
			WithAuth(&http.BasicAuth{
				Username: "root",
				Password: "password",
			}),
		},
			opts...)...,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = source.Close()
	})

	return source, false
}

func testStoreWithError(t *testing.T, gitRepoURL string, opts ...containers.Option[SnapshotStore]) error {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	source, err := NewSnapshotStore(ctx, zaptest.NewLogger(t), gitRepoURL,
		append([]containers.Option[SnapshotStore]{
			WithRef("main"),
			WithAuth(&http.BasicAuth{
				Username: "root",
				Password: "password",
			}),
		},
			opts...)...,
	)
	if err != nil {
		return err
	}

	t.Cleanup(func() {
		_ = source.Close()
	})

	return nil
}
