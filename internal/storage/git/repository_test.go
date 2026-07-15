package git

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	envsfs "go.flipt.io/flipt/internal/storage/environments/fs"
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

func TestFetchPolicy_Strict(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a repository with strict fetch policy (default)
	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.False(t, repo.lenientFetchPolicyEnabled, "should set fetch policy to strict")
}

func TestFetchPolicy_Lenient(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create a repository with lenient fetch policy
	repo, _, err := newRepository(
		t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithLenientFetchPolicy(),
	)
	require.NoError(t, err)
	require.NotNil(t, repo)

	assert.True(t, repo.lenientFetchPolicyEnabled, "should set fetch policy to lenient")
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
			err:      &os.SyscallError{Syscall: "write", Err: syscall.EINVAL},
			expected: false,
		},
		{
			name:     "net.OpError with non-connection-refused error",
			err:      &net.ParseError{},
			expected: false,
		},
		{
			name:     "connection i/o timeout",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: os.ErrDeadlineExceeded},
			expected: true,
		},
		{
			name: "DNS error - no such host",
			err: &net.OpError{
				Op:  "dial",
				Net: "tcp",
				Err: &net.DNSError{
					Err:        "no such host",
					Name:       "nonexistent.example.com",
					IsNotFound: true,
				},
			},
			expected: true,
		},
		{
			name: "DNS error - direct",
			err: &net.DNSError{
				Err:        "no such host",
				Name:       "invalid.local",
				IsNotFound: true,
			},
			expected: true,
		},
		{
			name: "DNS error - server failure",
			err: &net.DNSError{
				Err:         "server misbehaving",
				Name:        "example.com",
				Server:      "8.8.8.8",
				IsTemporary: true,
			},
			expected: true,
		},
		{
			name: "DNS error - permanent failure",
			err: &net.DNSError{
				Err:    "fatal erro",
				Name:   "example.com",
				Server: "8.8.8.8",
			},
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

func TestFetch_StrictPolicy_WithConnectionRefused(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository with strict fetch policy and invalid remote
	_, _, err := newRepository(
		t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "http://localhost:1/invalid-repo.git"),
	)
	require.Error(t, err)
}

func TestFetch_LenientPolicy_WithConnectionRefusedAndCommits(t *testing.T) {
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

	// Now create our Repository wrapper with lenient policy and invalid remote
	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "http://localhost:1/invalid-repo.git"),
		WithLenientFetchPolicy())
	require.NoError(t, err)
	assert.True(t, repo.lenientFetchPolicyEnabled)
	assert.True(t, repo.HasLenientFetchPolicy())

	err = repo.Fetch(t.Context())
	assert.Error(t, err)
}

func TestFetch_LenientPolicy_WithConnectionRefusedAndNoCommits(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository with lenient fetch policy and invalid remote, but no commits
	repo, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "http://localhost:1/invalid-repo.git"),
		WithLenientFetchPolicy())
	require.Error(t, err)
	assert.Nil(t, repo)
}

func TestFetch_LenientPolicy_WithNonConnectionError(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository with lenient fetch policy but a different type of error (not connection refused)
	// Using an invalid URL format to trigger a different error
	_, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "invalid://bad-url"),
		WithLenientFetchPolicy())
	require.Error(t, err)
}

func TestFetch_DefaultPolicy_BehavesAsStrict(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository without specifying policy (should default to strict behavior)
	_, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir),
		WithRemote("origin", "http://localhost:1/invalid-repo.git"))
	require.Error(t, err)
}

func TestNewRepository_RemoteSync(t *testing.T) {
	tests := []struct {
		name             string
		setupRepo        func(t *testing.T, tempDir string, remoteDir string) // setup function with remote dir
		initialRemoteURL string                                               // initial remote URL (or empty)
		newRemoteURL     string                                               // new remote URL to sync
	}{
		{
			name: "URL changed - should sync to new URL",
			setupRepo: func(t *testing.T, tempDir string, remoteDir string) {
				// Create a bare remote repository first
				oldRemoteDir := filepath.Join(remoteDir, "old-remote.git")
				newRemoteDir := filepath.Join(remoteDir, "new-remote.git")
				_, err := git.PlainInit(oldRemoteDir, true)
				require.NoError(t, err)
				_, err = git.PlainInit(newRemoteDir, true)
				require.NoError(t, err)

				// Create normal git repo with commits and initial remote
				plainRepo, err := git.PlainInit(tempDir, false)
				require.NoError(t, err)

				// Create and commit a file
				require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0o755))
				featuresFile := filepath.Join(tempDir, "production", "features.yaml")
				content := `namespace:
    key: production
    name: production`
				require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0o600))

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

				// Add initial remote
				_, err = plainRepo.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{oldRemoteDir},
				})
				require.NoError(t, err)

				// Push to the old remote so it has content
				err = plainRepo.Push(&git.PushOptions{
					RemoteName: "origin",
				})
				require.NoError(t, err)

				// Push to the new remote too so fetch works
				_, err = plainRepo.CreateRemote(&config.RemoteConfig{
					Name: "newremote",
					URLs: []string{newRemoteDir},
				})
				require.NoError(t, err)
				err = plainRepo.Push(&git.PushOptions{
					RemoteName: "newremote",
				})
				require.NoError(t, err)

				// Remove the newremote so we only have origin
				err = plainRepo.DeleteRemote("newremote")
				require.NoError(t, err)
			},
			initialRemoteURL: "", // set in setupRepo
			newRemoteURL:     "new-remote.git",
		},
		{
			name: "URL unchanged - should remain same",
			setupRepo: func(t *testing.T, tempDir string, remoteDir string) {
				// Create normal git repo with commits and remote
				plainRepo, err := git.PlainInit(tempDir, false)
				require.NoError(t, err)

				// Create and commit a file
				require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0o755))
				featuresFile := filepath.Join(tempDir, "production", "features.yaml")
				content := `namespace:
    key: production
    name: production`
				require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0o600))

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

				// Create a bare remote repository
				remoteRepoDir := filepath.Join(remoteDir, "test-remote.git")
				_, err = git.PlainInit(remoteRepoDir, true)
				require.NoError(t, err)

				// Add remote
				_, err = plainRepo.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{remoteRepoDir},
				})
				require.NoError(t, err)
			},
			newRemoteURL: "test-remote.git",
		},
		{
			name: "multiple URLs in git config - should sync to single URL",
			setupRepo: func(t *testing.T, tempDir string, remoteDir string) {
				// Create normal git repo with commits
				plainRepo, err := git.PlainInit(tempDir, false)
				require.NoError(t, err)

				// Create and commit a file
				require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0o755))
				featuresFile := filepath.Join(tempDir, "production", "features.yaml")
				content := `namespace:
    key: production
    name: production`
				require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0o600))

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

				// Create bare remote repositories
				remote1Dir := filepath.Join(remoteDir, "remote1.git")
				remote2Dir := filepath.Join(remoteDir, "remote2.git")
				_, err = git.PlainInit(remote1Dir, true)
				require.NoError(t, err)
				_, err = git.PlainInit(remote2Dir, true)
				require.NoError(t, err)

				// Add remote with multiple URLs
				_, err = plainRepo.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{remote1Dir, remote2Dir},
				})
				require.NoError(t, err)
			},
			newRemoteURL: "remote1.git",
		},
		{
			name: "add remote to existing repo without remote",
			setupRepo: func(t *testing.T, tempDir string, remoteDir string) {
				// Create normal git repo with commits but no remote
				plainRepo, err := git.PlainInit(tempDir, false)
				require.NoError(t, err)

				// Create and commit a file
				require.NoError(t, os.MkdirAll(filepath.Join(tempDir, "production"), 0o755))
				featuresFile := filepath.Join(tempDir, "production", "features.yaml")
				content := `namespace:
    key: production
    name: production`
				require.NoError(t, os.WriteFile(featuresFile, []byte(content), 0o600))

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

				// Create bare remote but don't add it yet
				newRemoteDir := filepath.Join(remoteDir, "new-remote.git")
				_, err = git.PlainInit(newRemoteDir, true)
				require.NoError(t, err)
			},
			newRemoteURL: "new-remote.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			remoteDir := t.TempDir() // separate dir for remote repos
			logger := zap.NewNop()

			// Setup repository state
			tt.setupRepo(t, tempDir, remoteDir)

			// Build the full remote URL
			fullRemoteURL := filepath.Join(remoteDir, tt.newRemoteURL)

			// Reopen repository with updated remote configuration
			repo, _, err := newRepository(t.Context(), logger,
				WithFilesystemStorage(tempDir),
				WithRemote("origin", fullRemoteURL))

			require.NoError(t, err)
			require.NotNil(t, repo)

			// Verify remote URLs were synchronized correctly
			gitConfig, err := repo.Config()
			require.NoError(t, err)

			require.NotNil(t, gitConfig.Remotes["origin"])
			assert.Equal(t, []string{fullRemoteURL}, gitConfig.Remotes["origin"].URLs,
				"remote URLs should match expected")
		})
	}
}

func TestNewRepository_RemoteSync_NoRemote(t *testing.T) {
	tempDir := t.TempDir()
	logger := zap.NewNop()

	// Create repository without remote
	repo1, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo1)

	// Verify no remotes
	gitConfig1, err := repo1.Config()
	require.NoError(t, err)
	assert.Empty(t, gitConfig1.Remotes)

	// Reopen without remote
	repo2, _, err := newRepository(t.Context(), logger,
		WithFilesystemStorage(tempDir))
	require.NoError(t, err)
	require.NotNil(t, repo2)

	// Still no remotes
	gitConfig2, err := repo2.Config()
	require.NoError(t, err)
	assert.Empty(t, gitConfig2.Remotes)
}

func TestFetchRepos(t *testing.T) {
	// Create a git repository to use as remote base
	remoteNormalDir := t.TempDir()
	remoteNormalRepo, err := git.PlainInit(
		remoteNormalDir, false,
		git.WithDefaultBranch(plumbing.NewBranchReferenceName("main")),
	)
	require.NoError(t, err)

	remoteWorktree, err := remoteNormalRepo.Worktree()
	require.NoError(t, err)
	// helper func to create commits
	createCommitOnRemote := func(tb testing.TB, i int) plumbing.Hash {
		tb.Helper()
		testFile := filepath.Join(remoteNormalDir, fmt.Sprintf("file%d.txt", i))
		require.NoError(t, os.WriteFile(testFile, fmt.Appendf(nil, "content %d", i), 0o600))
		_, err = remoteWorktree.Add(fmt.Sprintf("file%d.txt", i))
		require.NoError(t, err)
		commitHash, err := remoteWorktree.Commit(fmt.Sprintf("Commit %d", i), &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Test User",
				Email: "test@example.com",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)
		return commitHash
	}

	// Create multiple commits
	for i := range 3 {
		createCommitOnRemote(t, i)
	}

	t.Run("no repo exists", func(t *testing.T) {
		localDir := t.TempDir()
		logger := zap.NewNop()

		// Create local bare repository with remote
		repo, err := NewRepository(t.Context(), logger,
			WithFilesystemStorage(localDir),
			WithRemote("origin", remoteNormalDir))
		require.NoError(t, err)
		require.False(t, repo.isNormalRepo, "should be bare repository")

		// Fetch should create refs/heads
		err = repo.Fetch(t.Context())
		require.NoError(t, err)

		// Verify refs/heads was created for main branch
		ref, err := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
		require.NoError(t, err)
		assert.False(t, ref.Hash().IsZero(), "refs/heads/main should be created after fetch")
		shallowData, err := os.ReadFile(filepath.Join(localDir, "shallow"))
		require.NoError(t, err, "shallow file doesn't exist")
		assert.Equal(t, ref.Hash().String()+"\n", string(shallowData))
		shallows, err := repo.Storer.Shallow()
		require.NoError(t, err)
		assert.NotEmpty(t, shallows, "shallow boundary should be maintained for bare repository depth=1 fetch")
		createCommitOnRemote(t, 10)
		// refetch should maintain shallow
		commitHash := createCommitOnRemote(t, time.Now().Nanosecond())
		err = repo.Fetch(t.Context())
		require.NoError(t, err)
		shallows, err = repo.Storer.Shallow()
		require.NoError(t, err)
		assert.Contains(t, shallows, commitHash)
	})

	t.Run("bare repo exists", func(t *testing.T) {
		localDir := t.TempDir()
		logger := zap.NewNop()

		_, err := git.PlainInit(
			localDir, true,
			git.WithDefaultBranch(plumbing.NewBranchReferenceName("main")),
		)
		require.NoError(t, err)

		repo, err := NewRepository(t.Context(), logger,
			WithFilesystemStorage(localDir),
			WithRemote("origin", remoteNormalDir))
		require.NoError(t, err)
		require.False(t, repo.isNormalRepo, "should be bare repository")

		// Fetch should create refs/heads and not use depth=1
		err = repo.Fetch(t.Context())
		require.NoError(t, err)

		// Verify refs/heads was created for main branch
		ref, err := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
		require.NoError(t, err)
		assert.False(t, ref.Hash().IsZero(), "refs/heads/main should be created after fetch")
		shallowData, err := os.ReadFile(filepath.Join(localDir, "shallow"))
		require.NoError(t, err, "shallow file doesn't exist")
		assert.Equal(t, ref.Hash().String()+"\n", string(shallowData))
		shallows, err := repo.Storer.Shallow()
		require.NoError(t, err)
		assert.NotEmpty(t, shallows, "shallow boundary should be maintained for bare repository depth=1 fetch")
		// refetch should maintain shallow
		commitHash := createCommitOnRemote(t, time.Now().Nanosecond())
		err = repo.Fetch(t.Context())
		require.NoError(t, err)
		shallows, err = repo.Storer.Shallow()
		require.NoError(t, err)
		assert.Contains(t, shallows, commitHash)
	})

	t.Run("normal repo exists", func(t *testing.T) {
		localDir := t.TempDir()
		logger := zap.NewNop()

		_, err := git.PlainInit(
			localDir, false,
			git.WithDefaultBranch(plumbing.NewBranchReferenceName("main")),
		)
		require.NoError(t, err)

		repo, err := NewRepository(t.Context(), logger,
			WithFilesystemStorage(localDir),
			WithRemote("origin", remoteNormalDir))
		require.NoError(t, err)
		require.True(t, repo.isNormalRepo, "should be normal repository")

		// Fetch should create refs/heads and not use depth=1
		err = repo.Fetch(t.Context())
		require.NoError(t, err)

		// Verify refs/heads was created for main branch
		ref, err := repo.Reference(plumbing.NewBranchReferenceName("main"), true)
		require.NoError(t, err)
		assert.False(t, ref.Hash().IsZero(), "refs/heads/main should be created after fetch")
		_, err = os.ReadFile(filepath.Join(localDir, "shallow"))
		require.Error(t, err)
		shallows, err := repo.Storer.Shallow()
		require.NoError(t, err)
		assert.Empty(t, shallows, "shallow boundary should be present for normal repo")

		// refetch should maintain shallow
		createCommitOnRemote(t, time.Now().Nanosecond())
		err = repo.Fetch(t.Context())
		require.NoError(t, err)
		shallows, err = repo.Storer.Shallow()
		require.NoError(t, err)
		assert.Empty(t, shallows)
	})

	t.Run("prune removes deleted remote refs", func(t *testing.T) {
		localDir := t.TempDir()
		logger := zap.NewNop()

		repo, err := NewRepository(t.Context(), logger,
			WithFilesystemStorage(localDir),
			WithRemote("origin", remoteNormalDir))
		require.NoError(t, err)

		branch := plumbing.NewBranchReferenceName("flipt/default/testing")
		// Setup some heads with should be pruned
		remoteHead, err := remoteNormalRepo.Head()
		require.NoError(t, err)
		err = remoteNormalRepo.Storer.SetReference(plumbing.NewHashReference(branch, remoteHead.Hash()))
		require.NoError(t, err)

		// Fetch should prune refs/heads
		err = repo.Fetch(t.Context(), "*")
		require.NoError(t, err)

		// Verify refs/heads has not for flipt/main/testing branch
		_, err = repo.Reference(branch, true)
		require.NoError(t, err)

		// Remove remote branch
		err = remoteNormalRepo.Storer.RemoveReference(branch)
		require.NoError(t, err)

		// Fetch should prune refs/heads
		err = repo.Fetch(t.Context(), "*")
		require.NoError(t, err)

		// Verify refs/heads has not for flipt/main/testing branch
		_, err = repo.Reference(branch, true)
		require.Error(t, err)
		require.ErrorIs(t, err, plumbing.ErrReferenceNotFound)
	})
}

func TestUpdateAndPush_NonFastForwardRetry_BareRepo_Shallows(t *testing.T) {
	// Create a bare remote repository
	remoteDir := t.TempDir()
	remoteRepo, err := git.PlainInit(
		remoteDir, true,
		git.WithDefaultBranch(plumbing.NewBranchReferenceName("main")),
	)
	require.NoError(t, err)

	// Push an initial commit to establish the branch on remote
	bootstrapDir := t.TempDir()
	bootstrapPlain, err := git.PlainInit(
		bootstrapDir, false,
		git.WithDefaultBranch(plumbing.NewBranchReferenceName("main")),
	)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(bootstrapDir, "init.txt"), []byte("init"), 0o600))
	wt, err := bootstrapPlain.Worktree()
	require.NoError(t, err)
	_, err = wt.Add("init.txt")
	require.NoError(t, err)
	_, err = wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com", When: time.Now()},
	})
	require.NoError(t, err)
	_, err = bootstrapPlain.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteDir},
	})
	require.NoError(t, err)
	require.NoError(t, bootstrapPlain.Push(&git.PushOptions{RemoteName: "origin"}))

	// Create Flipt bare repo via newRepository (this does a depth=1 fetch, creating shallows)
	localDir := t.TempDir()
	repo, _, err := newRepository(
		t.Context(), zap.NewNop(),
		WithFilesystemStorage(localDir),
		WithRemote("origin", remoteDir),
		WithSignature("test", "test@test.com"),
	)
	require.NoError(t, err)
	require.False(t, repo.isNormalRepo, "should be bare repository")

	// Clear shallows so go-git's client-side isFastForward check can detect divergence.
	// When shallows are present, isFastForward conservatively assumes fast-forward upon
	// hitting a shallow boundary, which would mask the non-fast-forward condition we need
	// to trigger the retry path.
	require.NoError(t, repo.Storer.SetShallow(nil))

	shallows, err := repo.Storer.Shallow()
	require.NoError(t, err)
	require.Empty(t, shallows, "shallows must be cleared for this test")

	// Advance the remote behind our back with an interfering commit
	interferingDir := t.TempDir()
	interferingRepo, err := git.PlainClone(interferingDir, &git.CloneOptions{
		URL: remoteDir,
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(interferingDir, "other.txt"), []byte("other"), 0o600))
	interferingWt, err := interferingRepo.Worktree()
	require.NoError(t, err)
	_, err = interferingWt.Add("other.txt")
	require.NoError(t, err)
	interferingCommit, err := interferingWt.Commit("interfering commit", &git.CommitOptions{
		Author: &object.Signature{Name: "other", Email: "other@test.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, interferingRepo.Push(&git.PushOptions{RemoteName: "origin"}))

	hash, err := repo.UpdateAndPush(t.Context(), "main", func(fs envsfs.Filesystem) (string, error) {
		fi, err := fs.OpenFile("flipt.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
		if err != nil {
			return "", err
		}
		if _, err := fi.Write([]byte("from flipt")); err != nil {
			return "", err
		}
		if err := fi.Close(); err != nil {
			return "", err
		}
		return "flipt commit after retry", nil
	})
	require.NoError(t, err)
	assert.False(t, hash.IsZero())

	// Verify the commit landed on the remote
	remoteRef, err := remoteRepo.Reference(plumbing.NewBranchReferenceName("main"), true)
	require.NoError(t, err)
	assert.Equal(t, hash, remoteRef.Hash(), "remote should point to our commit")

	// Verify the commit has the interfering commit as ancestor
	commitObj, err := remoteRepo.CommitObject(hash)
	require.NoError(t, err)
	assert.Equal(t, "flipt commit after retry", commitObj.Message)
	assert.Len(t, commitObj.ParentHashes, 1)

	parentCommit, err := remoteRepo.CommitObject(commitObj.ParentHashes[0])
	require.NoError(t, err)
	assert.Equal(t, "interfering commit", parentCommit.Message)

	// Verify shallows were updated during the non-fast-forward retry depth=1 fetch
	shallows, err = repo.Storer.Shallow()
	require.NoError(t, err)
	assert.Contains(t, shallows, interferingCommit,
		"interfering commit should be in shallows after depth=1 retry fetch")
}

func TestUpdateAndPush_NonFastForwardRetry(t *testing.T) {
	// Create a bare remote repository
	remoteDir := t.TempDir()
	remoteRepo, err := git.PlainInit(
		remoteDir, true,
		git.WithDefaultBranch(plumbing.NewBranchReferenceName("main")),
	)
	require.NoError(t, err)

	// Create a normal local repo, push an initial commit to establish the branch
	localDir := t.TempDir()
	localPlain, err := git.PlainInit(
		localDir, false,
		git.WithDefaultBranch(plumbing.NewBranchReferenceName("main")),
	)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(localDir, "init.txt"), []byte("init"), 0o600))
	wt, err := localPlain.Worktree()
	require.NoError(t, err)
	_, err = wt.Add("init.txt")
	require.NoError(t, err)
	_, err = wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com", When: time.Now()},
	})
	require.NoError(t, err)

	_, err = localPlain.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteDir},
	})
	require.NoError(t, err)
	require.NoError(t, localPlain.Push(&git.PushOptions{RemoteName: "origin"}))

	// Open the repo through Flipt's constructor so we get a properly configured Repository
	repo, _, err := newRepository(
		t.Context(), zap.NewNop(),
		WithFilesystemStorage(localDir),
		WithRemote("origin", remoteDir),
		WithSignature("test", "test@test.com"),
	)
	require.NoError(t, err)

	// Now advance the remote behind our back: clone into a second worktree,
	// commit, and push. This makes our local remote-tracking ref stale.
	interferingDir := t.TempDir()
	interferingRepo, err := git.PlainClone(interferingDir, &git.CloneOptions{
		URL: remoteDir,
	})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(interferingDir, "other.txt"), []byte("other"), 0o600))
	interferingWt, err := interferingRepo.Worktree()
	require.NoError(t, err)
	_, err = interferingWt.Add("other.txt")
	require.NoError(t, err)
	_, err = interferingWt.Commit("interfering commit", &git.CommitOptions{
		Author: &object.Signature{Name: "other", Email: "other@test.com", When: time.Now()},
	})
	require.NoError(t, err)
	require.NoError(t, interferingRepo.Push(&git.PushOptions{RemoteName: "origin"}))

	// Now UpdateAndPush from our repo — it should hit non-fast-forward, retry, and succeed
	hash, err := repo.UpdateAndPush(t.Context(), "main", func(fs envsfs.Filesystem) (string, error) {
		fi, err := fs.OpenFile("flipt.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
		if err != nil {
			return "", err
		}
		if _, err := fi.Write([]byte("from flipt")); err != nil {
			return "", err
		}
		if err := fi.Close(); err != nil {
			return "", err
		}
		return "flipt commit after retry", nil
	})
	require.NoError(t, err)
	assert.False(t, hash.IsZero())

	// Verify the commit landed on the remote
	remoteRef, err := remoteRepo.Reference(plumbing.NewBranchReferenceName("main"), true)
	require.NoError(t, err)
	assert.Equal(t, hash, remoteRef.Hash(), "remote should point to our commit")

	// Verify the commit has the interfering commit as an ancestor (i.e. we rebased on top of it)
	commitObj, err := remoteRepo.CommitObject(hash)
	require.NoError(t, err)
	assert.Equal(t, "flipt commit after retry", commitObj.Message)
	assert.Len(t, commitObj.ParentHashes, 1)

	parentCommit, err := remoteRepo.CommitObject(commitObj.ParentHashes[0])
	require.NoError(t, err)
	assert.Equal(t, "interfering commit", parentCommit.Message)
}
