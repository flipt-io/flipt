package git

import (
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestStaticResolver(t *testing.T) {
	t.Run("should resolve static references correctly", func(t *testing.T) {
		repo := newGitRepo(t)
		resolver := staticResolver()

		commitHash := repo.createCommit(t)
		resolvedHash, err := resolver(repo.repo, "main")

		require.NoError(t, err)
		require.Equal(t, commitHash, resolvedHash)

		repo.checkout(t, "new-branch")

		commitHash = repo.createCommit(t)
		resolvedHash, err = resolver(repo.repo, "new-branch")

		require.NoError(t, err)
		require.Equal(t, commitHash, resolvedHash)
	})
}

func TestSemverResolver(t *testing.T) {
	t.Run("should resolve semver tags correctly when the reference is a constraint", func(t *testing.T) {
		repo := newGitRepo(t)
		resolver := semverResolver()
		constraint := "v0.1.*"

		commitHash := repo.createCommit(t)
		repo.createTag(t, "v0.1.0", commitHash)

		resolvedHash, err := resolver(repo.repo, constraint)

		require.NoError(t, err)
		require.Equal(t, commitHash, resolvedHash)

		// When committing again and creating a new tag respecting the constraint should resolve nicely
		commitHash = repo.createCommit(t)
		repo.createTag(t, "v0.1.4", commitHash)

		resolvedHash, err = resolver(repo.repo, constraint)

		require.NoError(t, err)
		require.Equal(t, commitHash, resolvedHash)
	})

	t.Run("should resolve semver tags correctly when the reference is not a constraint", func(t *testing.T) {
		repo := newGitRepo(t)
		resolver := semverResolver()

		commitHash := repo.createCommit(t)
		repo.createTag(t, "v0.1.0", commitHash)

		resolvedHash, err := resolver(repo.repo, "v0.1.0")

		require.NoError(t, err)
		require.Equal(t, commitHash, resolvedHash)
	})

	t.Run("should resolve semver tags correctly when there is non compliant semver tags", func(t *testing.T) {
		repo := newGitRepo(t)
		resolver := semverResolver()

		commitHash := repo.createCommit(t)
		repo.createTag(t, "non-semver-tag", commitHash)

		commitHash = repo.createCommit(t)
		repo.createTag(t, "v0.1.0", commitHash)

		resolvedHash, err := resolver(repo.repo, "v0.1.0")

		require.NoError(t, err)
		require.Equal(t, commitHash, resolvedHash)
	})

	t.Run("should return an error when no matching tag was found", func(t *testing.T) {
		repo := newGitRepo(t)
		resolver := semverResolver()

		commitHash := repo.createCommit(t)
		repo.createTag(t, "v0.1.0", commitHash)

		_, err := resolver(repo.repo, "v1.*.*")

		require.ErrorContains(t, err, "could not find the specified tag reference")
	})
}

type gitRepoTest struct {
	repo *git.Repository
}

func (g *gitRepoTest) createCommit(t *testing.T) plumbing.Hash {
	t.Helper()

	workTree, err := g.repo.Worktree()
	require.NoError(t, err)

	fileName := uuid.NewString()

	file, err := workTree.Filesystem.Create(fileName)
	require.NoError(t, err)
	defer file.Close()

	_, err = workTree.Add(fileName)
	require.NoError(t, err)

	commitHash, err := workTree.Commit(fmt.Sprintf("adding %s", fileName), &git.CommitOptions{
		Author: &object.Signature{Email: "dev@flipt.io", Name: "dev"},
	})
	require.NoError(t, err)

	return commitHash
}

func (g *gitRepoTest) checkout(t *testing.T, branchName string) {
	t.Helper()

	workTree, err := g.repo.Worktree()
	require.NoError(t, err)

	err = workTree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branchName)),
		Create: true,
	})
	require.NoError(t, err)
}

func (g *gitRepoTest) createTag(t *testing.T, tag string, commit plumbing.Hash) {
	t.Helper()

	_, err := g.repo.CreateTag(tag, commit, nil)
	require.NoError(t, err)
}

func newGitRepo(t *testing.T) *gitRepoTest {
	t.Helper()

	repo, err := git.InitWithOptions(memory.NewStorage(), memfs.New(), git.InitOptions{
		DefaultBranch: "refs/heads/main",
	})
	require.NoError(t, err)

	return &gitRepoTest{
		repo: repo,
	}
}
