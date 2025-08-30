package git

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewRepository_WithExistingFiles(t *testing.T) {
	t.Run("initialize repository with existing features.yaml", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir := t.TempDir()

		// Create a features.yaml file in the directory
		featuresPath := filepath.Join(tempDir, "features.yaml")
		featuresContent := []byte(`namespace:
  key: "default"
  name: "Default"
flags:
  - key: "test_flag"
    name: "Test Flag"
    type: "BOOLEAN_FLAG"
    enabled: true
`)
		err := os.WriteFile(featuresPath, featuresContent, 0600)
		require.NoError(t, err)

		// Create logger
		logger := zaptest.NewLogger(t)

		// Create repository with the directory containing features.yaml
		repo, err := NewRepository(
			context.Background(),
			logger,
			WithFilesystemStorage(tempDir),
		)
		require.NoError(t, err)
		require.NotNil(t, repo)

		// Verify that the repository was created successfully
		assert.NotNil(t, repo.Repository)

		// Check that the repository has a HEAD reference
		head, err := repo.Head()
		require.NoError(t, err)
		assert.NotNil(t, head)

		// Verify that the initial commit contains the features.yaml file
		commit, err := repo.CommitObject(head.Hash())
		require.NoError(t, err)
		assert.Contains(t, commit.Message, "Initial commit")

		// Check that the features.yaml file is in the commit
		tree, err := commit.Tree()
		require.NoError(t, err)

		file, err := tree.File("features.yaml")
		require.NoError(t, err)
		assert.NotNil(t, file)

		// Verify the content of the file
		content, err := file.Contents()
		require.NoError(t, err)
		assert.Equal(t, string(featuresContent), content)
	})

	t.Run("initialize empty repository when directory is empty", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir := t.TempDir()

		// Create logger
		logger := zaptest.NewLogger(t)

		// Create repository with empty directory
		repo, err := NewRepository(
			context.Background(),
			logger,
			WithFilesystemStorage(tempDir),
		)
		require.NoError(t, err)
		require.NotNil(t, repo)

		// Verify that the repository was created successfully
		assert.NotNil(t, repo.Repository)

		// Should have initial README commit
		head, err := repo.Head()
		require.NoError(t, err)
		assert.NotNil(t, head)

		commit, err := repo.CommitObject(head.Hash())
		require.NoError(t, err)
		assert.Contains(t, commit.Message, "add initial README")
	})

	t.Run("open existing git repository", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir := t.TempDir()

		// Initialize a git repository manually
		_, err := git.PlainInit(tempDir, false)
		require.NoError(t, err)

		// Create logger
		logger := zaptest.NewLogger(t)

		// Create repository with existing git repository
		repo, err := NewRepository(
			context.Background(),
			logger,
			WithFilesystemStorage(tempDir),
		)
		require.NoError(t, err)
		require.NotNil(t, repo)

		// Verify that the repository was opened successfully
		assert.NotNil(t, repo.Repository)
	})

	t.Run("initialize repository with multiple existing files", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir := t.TempDir()

		// Create multiple files in the directory
		files := map[string]string{
			"features.yaml": `namespace:
  key: "default"
flags:
  - key: "flag1"
`,
			"production/features.yaml": `namespace:
  key: "production"
flags:
  - key: "prod_flag"
`,
			"config.yml": `version: "1.0"
`,
		}

		// Create directories and files
		err := os.MkdirAll(filepath.Join(tempDir, "production"), 0755)
		require.NoError(t, err)

		for path, content := range files {
			fullPath := filepath.Join(tempDir, path)
			err := os.WriteFile(fullPath, []byte(content), 0600)
			require.NoError(t, err)
		}

		// Create logger
		logger := zaptest.NewLogger(t)

		// Create repository with the directory containing multiple files
		repo, err := NewRepository(
			context.Background(),
			logger,
			WithFilesystemStorage(tempDir),
		)
		require.NoError(t, err)
		require.NotNil(t, repo)

		// Verify that the repository was created successfully
		assert.NotNil(t, repo.Repository)

		// Check that the repository has a HEAD reference
		head, err := repo.Head()
		require.NoError(t, err)
		assert.NotNil(t, head)

		// Verify that the initial commit contains all the files
		commit, err := repo.CommitObject(head.Hash())
		require.NoError(t, err)
		assert.Contains(t, commit.Message, "Initial commit")

		// Check that all files are in the commit
		tree, err := commit.Tree()
		require.NoError(t, err)

		// Verify each file exists in the tree
		for path, expectedContent := range files {
			file, err := tree.File(path)
			require.NoError(t, err, "file %s should exist", path)
			assert.NotNil(t, file)

			// Verify the content of the file
			content, err := file.Contents()
			require.NoError(t, err)
			assert.Equal(t, expectedContent, content)
		}
	})
}
