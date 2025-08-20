package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/containers"
	"go.uber.org/zap/zaptest"
)

func TestNewRepository_LocalStorage(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	t.Run("creates new repository when directory is empty", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "flipt-repo-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tempDir) })

		opts := []containers.Option[Repository]{
			WithFilesystemStorage(tempDir),
		}

		repo, err := NewRepository(ctx, logger, opts...)
		require.NoError(t, err)
		assert.NotNil(t, repo)

		assert.NotNil(t, repo.Repository, "should create git repository")
	})

	t.Run("opens existing repository when directory has git repo", func(t *testing.T) {
		// This test specifically addresses the bug where git.Open(storage, nil)
		// would fail with "repository does not exist" for existing local repositories.
		// The fix uses git.PlainOpen() which properly handles local filesystem repos.

		tempDir, err := os.MkdirTemp("", "flipt-repo-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tempDir) })

		// Create a proper git repository first using PlainInit (simulating user-created repo)
		plainRepo, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		// Create a test file and commit it to make it a "real" repository
		testFile := filepath.Join(tempDir, "features.yml")
		err = os.WriteFile(testFile, []byte("namespace: default\nflags: []\n"), 0600)
		require.NoError(t, err)

		worktree, err := plainRepo.Worktree()
		require.NoError(t, err)

		_, err = worktree.Add("features.yml")
		require.NoError(t, err)

		_, err = worktree.Commit("Initial commit with features", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)

		// This is the critical test: Flipt should be able to open this existing repository
		// Before the fix, this would fail with "repository does not exist"
		opts := []containers.Option[Repository]{
			WithFilesystemStorage(tempDir),
		}

		repo, err := NewRepository(ctx, logger, opts...)
		require.NoError(t, err)
		assert.NotNil(t, repo)

		// Verify we can access the repository and it has the expected content
		assert.NotNil(t, repo.Repository, "should have access to underlying git repository")
	})

	t.Run("initializes git repo when directory has files but no git repo", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "flipt-repo-test-*")
		require.NoError(t, err)
		t.Cleanup(func() { os.RemoveAll(tempDir) })

		// Create multiple files in the directory (simulating user files that should be preserved)
		flagsFile := filepath.Join(tempDir, "features.yml")
		flagsContent := "namespace: default\nflags:\n  - key: test-flag\n    name: Test Flag\n    enabled: true\n"
		err = os.WriteFile(flagsFile, []byte(flagsContent), 0600)
		require.NoError(t, err)

		configFile := filepath.Join(tempDir, "config.yml")
		configContent := "version: v1\nnamespace: default\n"
		err = os.WriteFile(configFile, []byte(configContent), 0600)
		require.NoError(t, err)

		// Create a subdirectory with files
		subDir := filepath.Join(tempDir, "segments")
		err = os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		segmentFile := filepath.Join(subDir, "segments.yml")
		segmentContent := "namespace: default\nsegments: []\n"
		err = os.WriteFile(segmentFile, []byte(segmentContent), 0600)
		require.NoError(t, err)

		// Verify no .git directory exists initially
		_, err = os.Stat(filepath.Join(tempDir, ".git"))
		assert.True(t, os.IsNotExist(err), "should not have .git directory initially")

		opts := []containers.Option[Repository]{
			WithFilesystemStorage(tempDir),
		}

		// This should now successfully initialize a new repository automatically
		// when it detects files but no .git directory
		repo, err := NewRepository(ctx, logger, opts...)
		require.NoError(t, err)
		assert.NotNil(t, repo)

		// Verify a git repository was created
		_, err = os.Stat(filepath.Join(tempDir, ".git"))
		require.NoError(t, err, "should create .git directory automatically")

		// Verify we can access the repository
		assert.NotNil(t, repo.Repository, "should have access to initialized git repository")

		// Verify all original files are preserved
		preservedFlagsContent, err := os.ReadFile(flagsFile)
		require.NoError(t, err)
		assert.Equal(t, flagsContent, string(preservedFlagsContent), "original flags file should be preserved")

		preservedConfigContent, err := os.ReadFile(configFile)
		require.NoError(t, err)
		assert.Equal(t, configContent, string(preservedConfigContent), "original config file should be preserved")

		preservedSegmentContent, err := os.ReadFile(segmentFile)
		require.NoError(t, err)
		assert.Equal(t, segmentContent, string(preservedSegmentContent), "original segment file should be preserved")

		// Verify subdirectory structure is preserved
		_, err = os.Stat(subDir)
		require.NoError(t, err, "subdirectory should be preserved")

		// Verify the repository is functional - just check we can access the storer
		// The NewRepository function will handle adding the initial commit asynchronously
		assert.NotNil(t, repo.Storer, "repository should have a working storer")
	})

	t.Run("handles errors during auto-initialization gracefully", func(t *testing.T) {
		// This test verifies error handling when auto-initialization fails
		// Use an invalid path that would cause initialization to fail
		invalidPath := "/invalid/nonexistent/path/that/cannot/be/created"

		opts := []containers.Option[Repository]{
			WithFilesystemStorage(invalidPath),
		}

		// This should fail gracefully with a descriptive error
		repo, err := NewRepository(ctx, logger, opts...)
		require.Error(t, err, "should fail when path is invalid")
		assert.Nil(t, repo, "should not return repository on failure")

		// Just verify we get an error - the specific message may vary by OS
		assert.NotEmpty(t, err.Error(), "error should have a message")
	})
}

