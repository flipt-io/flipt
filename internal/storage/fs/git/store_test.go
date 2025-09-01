package git

import (
	"bytes"
	"context"
	"encoding/pem"
	"fmt"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
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

// RepoClone is a helper function to clone the git repository for test purposes,
// to be able to commit and push changes independently of the SnapshotStore fetch/poll logic.
func RepoClone(url string, workdir billy.Filesystem) (*git.Repository, error) {
	return git.Clone(memory.NewStorage(), workdir, &git.CloneOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		URL:        url,
		RemoteName: "origin",
	})
}

var gitRepoURL = os.Getenv("TEST_GIT_REPO_URL")
var gitRepoHEAD = os.Getenv("TEST_GIT_REPO_HEAD")
var gitRepoTAG = os.Getenv("TEST_GIT_REPO_TAG")

func Test_Store_String(t *testing.T) {
	require.Equal(t, "git", (&SnapshotStore{}).String())
}

func Test_Store_Subscribe_Hash(t *testing.T) {
	head := gitRepoHEAD
	if head == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_HEAD env var to run this test.")
		return
	}
	testStore(t, gitRepoURL, WithRef(head))
}

func Test_Store_View(t *testing.T) {
	ch := make(chan struct{})
	var once sync.Once
	store, skip := testStore(t, gitRepoURL, WithPollOptions(
		fs.WithInterval(time.Second),
		fs.WithNotify(t, func(modified bool) {
			if modified {
				once.Do(func() { close(ch) })
			}
		}),
	))
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Use a separate clone working repo for modifying and pushing changes
	workdir := memfs.New()
	repo, err := RepoClone(gitRepoURL, workdir)
	require.NoError(t, err)

	tree, err := repo.Worktree()
	require.NoError(t, err)

	// Checkout main branch explicitly
	require.NoError(t, tree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

	// Update features.yml in the memfs-backed repo
	fi, err := workdir.OpenFile("features.yml", os.O_TRUNC|os.O_RDWR, os.ModePerm)
	require.NoError(t, err)

	updated := []byte(`namespace: production
flags:
    - key: foo
      name: Foo`)

	_, err = fi.Write(updated)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	// Commit changes
	_, err = tree.Commit("chore: update features.yml", &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	// Push changes with BasicAuth
	err = repo.Push(&git.PushOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
	})
	require.NoError(t, err)

	// Wait until poll notifies snapshot updated or timeout
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for snapshot")
	}

	t.Log("received new snapshot")

	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err = s.GetFlag(ctx, storage.NewResource("production", "foo"))
		return err
	}))
}

func Test_Store_View_WithFilesystemStorage(t *testing.T) {
	dir := t.TempDir()

	for i := range []int{1, 2, 3} {
		i := i
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			ch := make(chan struct{})
			var once sync.Once
			store, skip := testStore(t, gitRepoURL,
				WithFilesystemStorage(dir),
				WithPollOptions(
					fs.WithInterval(time.Second),
					fs.WithNotify(t, func(modified bool) {
						if modified {
							once.Do(func() { close(ch) })
						}
					}),
				),
			)
			if skip {
				return
			}

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			// Use a separate clone to make changes to the repo
			workdir := memfs.New()
			repo, err := RepoClone(gitRepoURL, workdir)
			require.NoError(t, err)

			tree, err := repo.Worktree()
			require.NoError(t, err)

			require.NoError(t, tree.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewBranchReferenceName("main"),
			}))

			fi, err := workdir.OpenFile("features.yml", os.O_TRUNC|os.O_RDWR, os.ModePerm)
			require.NoError(t, err)

			updated := []byte(`namespace: production
flags:
    - key: foo
      name: Foo
      description: Foo description` + fmt.Sprintf(" %d", i))

			_, err = fi.Write(updated)
			require.NoError(t, err)
			require.NoError(t, fi.Close())

			_, err = tree.Commit("chore: update features.yml", &git.CommitOptions{
				All:    true,
				Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
			})
			require.NoError(t, err)

			err = repo.Push(&git.PushOptions{
				Auth:       &http.BasicAuth{Username: "root", Password: "password"},
				RemoteName: "origin",
			})
			require.NoError(t, err)

			select {
			case <-ch:
			case <-time.After(time.Minute):
				t.Fatal("timed out waiting for snapshot")
			}

			t.Log("received new snapshot")

			require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
				_, err = s.GetFlag(ctx, storage.NewResource("production", "foo"))
				return err
			}))
		})
	}
}

func Test_Store_View_WithRevision(t *testing.T) {
	ch := make(chan struct{})
	var once sync.Once
	store, skip := testStore(t, gitRepoURL, WithPollOptions(
		fs.WithInterval(time.Second),
		fs.WithNotify(t, func(modified bool) {
			if modified {
				once.Do(func() { close(ch) })
			}
		}),
	))
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Use separate clone to create and push changes on new branch
	workdir := memfs.New()
	repo, err := RepoClone(gitRepoURL, workdir)
	require.NoError(t, err)

	tree, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, tree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("new-branch"),
		Create: true,
	}))

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

	_, err = tree.Commit("chore: update features.yml add foo and bar", &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	err = repo.Push(&git.PushOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/new-branch:refs/heads/new-branch"},
	})
	require.NoError(t, err)

	// The "bar" flag should not exist on default revision "main"
	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "bar"))
		require.Error(t, err, "flag should not be found in default revision")
		return nil
	}))

	// Also not present on "main" explicitly
	require.NoError(t, store.View(ctx, "main", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "bar"))
		require.Error(t, err, "flag should not be found in explicitly named main revision")
		return nil
	}))

	// Should be present on new-branch
	require.NoError(t, store.View(ctx, "new-branch", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "bar"))
		require.NoError(t, err, "flag should be present on new-branch")
		return nil
	}))

	// "baz" flag should not exist yet
	require.NoError(t, store.View(ctx, "new-branch", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "baz"))
		require.Error(t, err, "flag should not be found in explicitly named new-branch revision")
		return nil
	}))

	// Add "baz" flag
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

	_, err = tree.Commit("chore: update features.yml add baz", &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	err = repo.Push(&git.PushOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/new-branch:refs/heads/new-branch"},
	})
	require.NoError(t, err)

	// Wait for poll to pick up new commit
	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for fetch")
	}

	// Now "baz" flag must be present on "new-branch"
	require.NoError(t, store.View(ctx, "new-branch", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("production", "baz"))
		require.NoError(t, err, "flag should be present on new-branch")
		return nil
	}))
}

func Test_Store_View_WithSemverRevision(t *testing.T) {
	tag := gitRepoTAG
	if tag == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_TAG env var to run this test.")
		return
	}

	head := gitRepoHEAD
	if head == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_HEAD env var to run this test.")
		return
	}

	ch := make(chan struct{})
	var once sync.Once
	store, skip := testStore(t, gitRepoURL,
		WithRef("v0.1.*"),
		WithSemverResolver(),
		WithPollOptions(
			fs.WithInterval(time.Second),
			fs.WithNotify(t, func(modified bool) {
				if modified {
					once.Do(func() { close(ch) })
				}
			}),
		),
	)
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	workdir := memfs.New()
	repo, err := RepoClone(gitRepoURL, workdir)
	require.NoError(t, err)

	tree, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, tree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("semver-branch"),
		Create: true,
	}))

	// Verify that bar flag absent on default revision
	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err := s.GetFlag(ctx, storage.NewResource("semver", "bar"))
		require.Error(t, err, "flag should not be found in default revision")
		return nil
	}))

	fi, err := workdir.OpenFile("features.yml", os.O_TRUNC|os.O_RDWR, os.ModePerm)
	require.NoError(t, err)

	updated := []byte(`namespace: semver
flags:
  - key: foo
    name: Foo
  - key: bar
    name: Bar`)

	_, err = fi.Write(updated)
	require.NoError(t, err)
	require.NoError(t, fi.Close())

	commit, err := tree.Commit("chore: update features.yml", &git.CommitOptions{
		All:    true,
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	_, err = repo.CreateTag("v0.1.4", commit, nil)
	require.NoError(t, err)

	err = repo.Push(&git.PushOptions{
		Auth:       &http.BasicAuth{Username: "root", Password: "password"},
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			"refs/heads/semver-branch:refs/heads/semver-branch",
			"refs/tags/v0.1.4:refs/tags/v0.1.4",
		},
	})
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(time.Minute):
		t.Fatal("timed out waiting for snapshot")
	}

	t.Log("received new snapshot")

	hash, err := store.Resolve("v0.1.*")
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
	var once sync.Once
	store, skip := testStore(t, gitRepoURL, WithPollOptions(
		fs.WithInterval(time.Second),
		fs.WithNotify(t, func(modified bool) {
			if modified {
				once.Do(func() { close(ch) })
			}
		}),
	),
		WithDirectory("subdir"),
	)
	if skip {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	workdir := memfs.New()
	repo, err := RepoClone(gitRepoURL, workdir)
	require.NoError(t, err)

	tree, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, tree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))

	require.NoError(t, store.View(ctx, "", func(s storage.ReadOnlyStore) error {
		_, err = s.GetFlag(ctx, storage.NewResource("alternative", "otherflag"))
		return err
	}))
}

func Test_Store_SelfSignedSkipTLS(t *testing.T) {
	ts := httptest.NewTLSServer(nil)
	defer ts.Close()

	err := testStoreWithError(t, ts.URL, WithInsecureTLS(false))
	require.ErrorContains(t, err, "tls: failed to verify certificate: x509: certificate signed by unknown authority")

	err = testStoreWithError(t, ts.URL, WithInsecureTLS(true))
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

	err = testStoreWithError(t, ts.URL)
	require.ErrorContains(t, err, "tls: failed to verify certificate: x509: certificate signed by unknown authority")

	err = testStoreWithError(t, ts.URL, WithCABundle(buf.Bytes()))
	require.ErrorIs(t, err, transport.ErrRepositoryNotFound)
}

func testStore(t *testing.T, gitRepoURL string, opts ...containers.Option[SnapshotStore]) (*SnapshotStore, bool) {
	t.Helper()
	if gitRepoURL == "" {
		t.Skip("Set non-empty TEST_GIT_REPO_URL env var to run this test.")
		return nil, true
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	store, err := NewSnapshotStore(ctx, zaptest.NewLogger(t), gitRepoURL,
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
		_ = store.Close()
	})

	return store, false
}

func testStoreWithError(t *testing.T, gitRepoURL string, opts ...containers.Option[SnapshotStore]) error {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	store, err := NewSnapshotStore(ctx, zaptest.NewLogger(t), gitRepoURL,
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
		_ = store.Close()
	})

	return nil
}
