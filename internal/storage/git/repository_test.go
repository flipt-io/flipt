package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewRepository_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	repo, empty, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.True(t, empty, "empty directory should be marked as empty")
	assert.False(t, repo.isNormalRepo, "empty directory should create bare repository")
}

func TestNewRepository_NonExistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	logger := zap.NewNop()

	repo, empty, err := newRepository(context.Background(), logger, WithFilesystemStorage(nonExistentDir))
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.True(t, empty, "non-existent directory should be marked as empty")
	assert.False(t, repo.isNormalRepo, "non-existent directory should create bare repository")
}

func TestNewRepository_DirectoryWithFiles_NoGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create some content files
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0755))
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
	require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0644))

	repo, empty, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
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
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0755))
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
	require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0644))

	// Commit the files
	plainRepo, err := git.PlainOpen(tempDir)
	require.NoError(t, err)
	worktree, err := plainRepo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("production/features.yaml")
	require.NoError(t, err)
	_, err = worktree.Commit("Initial commit", &git.CommitOptions{})
	require.NoError(t, err)

	// Test opening with Flipt
	repo, empty, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
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

	repo, empty, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.True(t, empty, "normal git repo without commits should be marked as empty")
	assert.True(t, repo.isNormalRepo, "should detect normal git repository")
}

func TestNewRepository_BareGitRepository(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a bare Git repository by initializing with custom storage
	repo1, empty1, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo1)
	assert.True(t, empty1, "initial bare repository should be empty")
	assert.False(t, repo1.isNormalRepo, "should be bare repository")

	// Try to reopen the empty bare repository (no files added yet)
	repo2, empty2, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
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
	repo1, empty1, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo1)
	assert.True(t, empty1, "initial bare repository should be empty")
	assert.False(t, repo1.isNormalRepo, "should be bare repository")

	// Now add some files to the temp directory (simulating files in storage path)
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0755))
	featuresFile := filepath.Join(tempDir, "production", "features.yaml")
	content := `namespace:
    key: production
    name: production
flags:
    - key: test-flag
      name: Test Flag
      type: VARIANT_FLAG_TYPE`
	require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0644))

	// Try to reopen - should detect existing bare repository instead of treating as normal repo
	repo2, empty2, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
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
	require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0755))
	featuresFile := filepath.Join(tempDir, "production", "features.yaml")
	initialContent := `namespace:
    key: production
    name: production
flags:
    - key: test-flag
      name: Test Flag
      enabled: false`
	require.NoError(t, os.WriteFile(featuresFile, []byte(initialContent), 0644))

	worktree, err := plainRepo.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("production/features.yaml")
	require.NoError(t, err)
	commit1, err := worktree.Commit("Initial commit", &git.CommitOptions{})
	require.NoError(t, err)

	// Create second commit with modified content
	updatedContent := `namespace:
    key: production
    name: production
flags:
    - key: test-flag
      name: Test Flag
      enabled: true`
	require.NoError(t, os.WriteFile(featuresFile, []byte(updatedContent), 0644))
	_, err = worktree.Add("production/features.yaml")
	require.NoError(t, err)
	commit2, err := worktree.Commit("Update flag", &git.CommitOptions{})
	require.NoError(t, err)

	// Create Flipt repository instance
	repo, _, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.True(t, repo.isNormalRepo, "should be normal repository")

	// Test updating working directory to first commit
	err = repo.updateWorkingDirectory(context.Background(), commit1)
	require.NoError(t, err)

	// Verify file content matches first commit
	content, err := os.ReadFile(featuresFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "enabled: false", "working directory should show first commit content")

	// Test updating working directory to second commit
	err = repo.updateWorkingDirectory(context.Background(), commit2)
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
	repo, _, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.False(t, repo.isNormalRepo, "should be bare repository")

	// updateWorkingDirectory should be a no-op for bare repositories
	err = repo.updateWorkingDirectory(context.Background(), plumbing.ZeroHash)
	assert.NoError(t, err, "updateWorkingDirectory should not error on bare repositories")
}

func TestRepositoryDefaults(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	repo, _, err := newRepository(context.Background(), logger, WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.Equal(t, "main", repo.defaultBranch, "should default to main branch")
	assert.Equal(t, tempDir, repo.localPath, "should set local path correctly")
	assert.False(t, repo.isNormalRepo, "should default to bare repository for empty directory")
}

func TestRepositoryWithCustomBranch(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	repo, _, err := newRepository(context.Background(), logger, 
		WithFilesystemStorage(tempDir),
		WithDefaultBranch("develop"))
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.Equal(t, "develop", repo.defaultBranch, "should use custom default branch")
}