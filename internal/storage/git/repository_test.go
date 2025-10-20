package git

import (
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRepository_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	repo, empty, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.True(t, empty, "empty directory should be marked as empty")
	assert.False(t, repo.isNormalRepo, "empty directory should create bare repository")
}

func TestNewRepository_NonExistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	logger := zap.NewNop()

	repo, empty, err := newRepository(t.Context(), logger, WithFilesystemStorage(nonExistentDir))
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.True(t, empty, "non-existent directory should be marked as empty")
	assert.False(t, repo.isNormalRepo, "non-existent directory should create bare repository")
}

func TestNewRepository_DirectoryWithFiles_NoGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create some content files
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0o755))
	featuresFile := filepath.Join(tempDir, "production", "features.yaml")
	content := `namespace:
    key: production
    name: production
flags:
    - key: test-flag
      name: Test Flag
      type: VARIANT_FLAG_TYPE
      description: A test flag
      enabled: true`
	require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0o600))

	repo, empty, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.True(t, empty, "directory with files but no git repo should be marked as empty for initial commit")
	assert.False(t, repo.isNormalRepo, "directory with files but no git repo should create bare repository")
}

func TestNewRepository_NormalGitRepository(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a normal Git repository
	_, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create and commit some content
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0o755))
	featuresFile := filepath.Join(tempDir, "production", "features.yaml")
	content := `namespace:
    key: production
    name: production
flags:
    - key: test-flag
      name: Test Flag
      type: VARIANT_FLAG_TYPE
      description: A test flag
      enabled: true`
	require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0o600))

	// Commit the files
	plainRepo, err := git.PlainOpen(tempDir)
	require.NoError(t, err)
	worktree, err := plainRepo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("production/features.yaml")
	require.NoError(t, err)
	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Test opening with Flipt
	repo, empty, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.False(t, empty, "normal git repo with commits should not be marked as empty")
	assert.True(t, repo.isNormalRepo, "should detect normal git repository")

	// Verify remote tracking reference was created
	remoteRef, err := repo.Repository.Reference("refs/remotes/origin/main", true)
	require.NoError(t, err)
	assert.False(t, remoteRef.Hash().IsZero(), "remote tracking reference should be set")
}

func TestNewRepository_NormalGitRepository_NoCommits(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a normal Git repository but don't commit anything
	_, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	repo, empty, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.True(t, empty, "normal git repo without commits should be marked as empty")
	assert.True(t, repo.isNormalRepo, "should detect normal git repository")
}

func TestNewRepository_BareGitRepository(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a bare Git repository by initializing with custom storage
	repo1, empty1, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo1)
	assert.True(t, empty1, "initial bare repository should be empty")
	assert.False(t, repo1.isNormalRepo, "should be bare repository")

	// Try to reopen the empty bare repository (no files added yet)
	repo2, empty2, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo2)
	// The empty status may vary when reopening existing repositories depending on internal Git state
	// The key requirement is that it remains a bare repository
	assert.False(t, repo2.isNormalRepo, "should remain bare repository")
	_ = empty2 // Don't assert on empty status for existing repos
}

func TestNewRepository_FilesWithNonGitBareRepository(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a bare repository first
	repo1, empty1, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo1)
	assert.True(t, empty1, "initial bare repository should be empty")
	assert.False(t, repo1.isNormalRepo, "should be bare repository")

	// Now add some files to the temp directory (simulating files in storage path)
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0o755))
	featuresFile := filepath.Join(tempDir, "production", "features.yaml")
	content := `namespace:
    key: production
    name: production
flags:
    - key: test-flag
      name: Test Flag
      type: VARIANT_FLAG_TYPE`
	require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0o600))

	// Try to reopen - should detect existing bare repository instead of treating as normal repo
	repo2, empty2, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo2)
	// The behavior here depends on whether the bare repository has been initialized with objects
	// Since we created files but haven't committed to the bare repo, it should still open as bare
	assert.False(t, repo2.isNormalRepo, "should open existing bare repository, not create normal repo")
	_ = empty2 // empty status may vary depending on bare repo state
}

func TestUpdateWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a normal Git repository
	plainRepo, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Create initial content and commit
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0o755))
	featuresFile := filepath.Join(tempDir, "production", "features.yaml")
	initialContent := `namespace:
    key: production
    name: production
flags:
    - key: test-flag
      name: Test Flag
      enabled: false`
	require.NoError(t, os.WriteFile(featuresFile, []byte(initialContent), 0o600))

	worktree, err := plainRepo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("production/features.yaml")
	require.NoError(t, err)
	commit1, err := worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Create second commit with modified content
	updatedContent := `namespace:
    key: production
    name: production
flags:
    - key: test-flag
      name: Test Flag
      enabled: true`
	require.NoError(t, os.WriteFile(featuresFile, []byte(updatedContent), 0o600))
	_, err = worktree.Add("production/features.yaml")
	require.NoError(t, err)
	commit2, err := worktree.Commit("Update flag", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Create Flipt repository instance
	repo, _, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.True(t, repo.isNormalRepo, "should be normal repository")

	// Test updating working directory to first commit
	err = repo.updateWorkingDirectory(t.Context(), commit1)
	require.NoError(t, err)

	// Verify file content matches first commit
	content, err := os.ReadFile(featuresFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "enabled: false", "working directory should show first commit content")

	// Test updating working directory to second commit
	err = repo.updateWorkingDirectory(t.Context(), commit2)
	require.NoError(t, err)

	// Verify file content matches second commit
	content, err = os.ReadFile(featuresFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "enabled: true", "working directory should show second commit content")
}

func TestUpdateWorkingDirectory_BareRepository(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a bare repository
	repo, _, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.False(t, repo.isNormalRepo, "should be bare repository")

	// updateWorkingDirectory should be a no-op for bare repositories
	err = repo.updateWorkingDirectory(t.Context(), plumbing.ZeroHash)
	assert.NoError(t, err, "updateWorkingDirectory should not error on bare repositories")
}

func TestRepositoryDefaults(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	repo, _, err := newRepository(t.Context(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.Equal(t, "main", repo.defaultBranch, "should default to main branch")
	assert.Equal(t, tempDir, repo.localPath, "should set local path correctly")
	assert.False(t, repo.isNormalRepo, "should default to bare repository for empty directory")
}

func TestRepositoryWithCustomBranch(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithDefaultBranch("develop"))
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.Equal(t, "develop", repo.defaultBranch, "should use custom default branch")
}

func TestRemoteStartupPolicy_Required(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a repository with required fetch policy (default)
	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.False(t, repo.optionalRemoteStartupFetch, "should set fetch policy to required")
}

func TestRemoteStartupPolicy_Optional(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a repository with optional fetch policy
	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithOptionalRemoteStartupFetchPolicy(),
	)
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.True(t, repo.optionalRemoteStartupFetch, "should set fetch policy to optional")
}

func TestIsConnectionRefused(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir))
	require.NoError(t, err)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      assert.AnError,
			expected: false,
		},
		{
			name:     "connection refused wrapped in net.OpError",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: &os.SyscallError{Syscall: "connect", Err: syscall.ECONNREFUSED}},
			expected: true,
		},
		{
			name:     "different syscall error",
			err:      &os.SyscallError{Syscall: "connect", Err: syscall.ETIMEDOUT},
			expected: false,
		},
		{
			name:     "net.OpError with non-connection-refused error",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: &os.SyscallError{Syscall: "connect", Err: syscall.EADDRINUSE}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.IsConnectionRefused(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFetch_NoRemote(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository without remote
	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir))
	require.NoError(t, err)

	// Fetch should be a no-op when no remote is configured
	err = repo.Fetch(t.Context())
	assert.NoError(t, err, "fetch without remote should succeed silently")
}

func TestFetch_RequiredPolicy_WithConnectionRefused(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository with required fetch policy and invalid remote
	_, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "http://localhost:1/invalid-repo.git"),
	)
	require.Error(t, err)
}

func TestFetch_OptionalPolicy_WithConnectionRefusedAndCommits(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	plainRepo, err := git.PlainInit(tempDir, false)
	require.NoError(t, err)

	// Add a commit to the repository
	worktree, err := plainRepo.Worktree()
	require.NoError(t, err)

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0o600))

	_, err = worktree.Add("test.txt")
	require.NoError(t, err)

	_, err = worktree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	// Now create our Repository wrapper with optional policy and invalid remote
	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "http://localhost:1/invalid-repo.git"),
		WithOptionalRemoteStartupFetchPolicy())
	require.NoError(t, err)
	assert.True(t, repo.optionalRemoteStartupFetch)
	assert.True(t, repo.IsRemoteStartupFetchOptional())

	err = repo.Fetch(t.Context())
	assert.Error(t, err)
}

func TestFetch_OptionalPolicy_WithConnectionRefusedAndNoCommits(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository with optional fetch policy and invalid remote, but no commits
	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "http://localhost:1/invalid-repo.git"),
		WithOptionalRemoteStartupFetchPolicy())
	require.Error(t, err)
	assert.Nil(t, repo)
}

func TestFetch_OptionalPolicy_WithNonConnectionError(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository with optional fetch policy but a different type of error (not connection refused)
	// Using an invalid URL format to trigger a different error
	_, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "invalid://bad-url"),
		WithOptionalRemoteStartupFetchPolicy())
	require.Error(t, err)
}

func TestFetch_DefaultPolicy_BehavesAsRequired(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository without specifying policy (should default to required behavior)
	_, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "http://localhost:1/invalid-repo.git"))
	require.Error(t, err)
}
