package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

func TestNewProvider(t *testing.T) {
	t.Run("checks base directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Cleanup(func() {
			os.RemoveAll(tmpDir)
		})

		provider, err := NewProvider(tmpDir, zap.NewNop())

		require.NoError(t, err)
		assert.Equal(t, tmpDir, provider.basePath)
	})

	t.Run("fails when base directory does not exist", func(t *testing.T) {
		basePath := "/root/invalid/path"

		_, err := NewProvider(basePath, zap.NewNop())

		require.Error(t, err)
		assert.Contains(t, err.Error(), "checking base path")
	})
}

func TestProvider_GetSecret(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	provider, err := NewProvider(tmpDir, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("retrieves existing secret", func(t *testing.T) {
		// Create a secret file directly (simplified approach)
		secretPath := "test_secret"
		secretValue := "my-secret-value"

		fullPath := filepath.Join(tmpDir, secretPath)
		err := os.WriteFile(fullPath, []byte(secretValue), 0600)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, secretPath)

		require.NoError(t, err)
		assert.Equal(t, secretPath, retrieved.Path)
		assert.Equal(t, map[string][]byte{secretPath: []byte(secretValue)}, retrieved.Data)
	})

	t.Run("retrieves nested secret", func(t *testing.T) {
		// Create nested directory structure
		secretPath := "azure/storage_key"        // #nosec G101 - test data
		secretValue := "azure-storage-key-value" // #nosec G101 - test data

		fullPath := filepath.Join(tmpDir, secretPath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0700)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(secretValue), 0600)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, secretPath)

		require.NoError(t, err)
		assert.Equal(t, secretPath, retrieved.Path)
		assert.Equal(t, map[string][]byte{secretPath: []byte(secretValue)}, retrieved.Data)
	})

	t.Run("trims whitespace and newlines", func(t *testing.T) {
		// Create a secret with trailing whitespace
		secretPath := "whitespace_test"
		secretValue := "  my-secret-value\n\n  " // #nosec G101 - test data
		expectedValue := "my-secret-value"

		fullPath := filepath.Join(tmpDir, secretPath)
		err := os.WriteFile(fullPath, []byte(secretValue), 0600)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, secretPath)

		require.NoError(t, err)
		assert.Equal(t, map[string][]byte{secretPath: []byte(expectedValue)}, retrieved.Data)
	})

	t.Run("returns error for non-existent secret", func(t *testing.T) {
		_, err := provider.GetSecret(ctx, "non/existent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found")
	})

	t.Run("handles GPG key with newlines", func(t *testing.T) {
		// Create a GPG key with newlines
		secretPath := "gpg_key"
		//nolint:gosec // this is a test
		gpgKey := "-----BEGIN PGP PRIVATE KEY BLOCK-----\n\nlQVYBGh9OVoBDACmOSHo...\n-----END PGP PRIVATE KEY BLOCK-----"

		fullPath := filepath.Join(tmpDir, secretPath)
		err := os.WriteFile(fullPath, []byte(gpgKey), 0600)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, secretPath)

		require.NoError(t, err)
		assert.Equal(t, []byte(gpgKey), retrieved.Data[secretPath])
	})
}

func TestProvider_PutSecret(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	provider, err := NewProvider(tmpDir, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("stores secret with correct permissions", func(t *testing.T) {
		secretPath := "test_secret"
		secretValue := "my-secret-value"
		secret := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				secretPath: []byte(secretValue),
			},
		}

		err := provider.PutSecret(ctx, secretPath, secret)

		require.NoError(t, err)

		// Verify file was created with correct permissions
		fullPath := provider.secretPath(secretPath)
		stat, err := os.Stat(fullPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), stat.Mode().Perm())

		// Verify content
		content, err := os.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, secretValue, string(content))
	})

	t.Run("stores nested secret", func(t *testing.T) {
		secretPath := "aws/prod/access_key"   // #nosec G101 - test data
		secretValue := "aws-access-key-value" // #nosec G101 - test data
		secret := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				secretPath: []byte(secretValue),
			},
		}

		err := provider.PutSecret(ctx, secretPath, secret)

		require.NoError(t, err)

		// Verify directory structure was created
		fullPath := provider.secretPath(secretPath)
		dir := filepath.Dir(fullPath)
		dirStat, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, dirStat.IsDir())
		assert.Equal(t, os.FileMode(0700), dirStat.Mode().Perm())

		// Verify content
		content, err := os.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, secretValue, string(content))
	})

	t.Run("overwrites existing secret", func(t *testing.T) {
		secretPath := "overwrite_test"

		// Create initial secret
		secret1 := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				secretPath: []byte("old_value"),
			},
		}

		err := provider.PutSecret(ctx, secretPath, secret1)
		require.NoError(t, err)

		// Overwrite with new secret
		secret2 := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				secretPath: []byte("new_value"),
			},
		}

		err = provider.PutSecret(ctx, secretPath, secret2)
		require.NoError(t, err)

		// Verify the secret was overwritten
		retrieved, err := provider.GetSecret(ctx, secretPath)
		require.NoError(t, err)
		assert.Equal(t, []byte("new_value"), retrieved.Data[secretPath])
	})

	t.Run("returns error when secret data missing for path", func(t *testing.T) {
		secretPath := "missing_data"
		secret := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				"different_key": []byte("value"),
			},
		}

		err := provider.PutSecret(ctx, secretPath, secret)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no secret value found for path")
	})
}

func TestProvider_DeleteSecret(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	provider, err := NewProvider(tmpDir, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("deletes existing secret", func(t *testing.T) {
		secretPath := "delete_test"
		secretValue := "delete-me"

		// Create the secret file directly
		fullPath := filepath.Join(tmpDir, secretPath)
		err := os.WriteFile(fullPath, []byte(secretValue), 0600)
		require.NoError(t, err)

		// Verify it exists
		_, err = provider.GetSecret(ctx, secretPath)
		require.NoError(t, err)

		// Delete it
		err = provider.DeleteSecret(ctx, secretPath)
		require.NoError(t, err)

		// Verify it's gone
		_, err = provider.GetSecret(ctx, secretPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found")
	})

	t.Run("returns error for non-existent secret", func(t *testing.T) {
		err := provider.DeleteSecret(ctx, "non/existent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found")
	})

	t.Run("cleans up empty directories", func(t *testing.T) {
		secretPath := "cleanup/test/deep/secret"

		// Create the secret file
		fullPath := filepath.Join(tmpDir, secretPath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0700)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte("test"), 0600)
		require.NoError(t, err)

		// Verify directories were created
		deepDir := filepath.Join(tmpDir, "cleanup", "test", "deep")
		_, err = os.Stat(deepDir)
		require.NoError(t, err)

		// Delete the secret
		err = provider.DeleteSecret(ctx, secretPath)
		require.NoError(t, err)

		// Verify empty directories were cleaned up
		_, err = os.Stat(deepDir)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("does not remove non-empty directories", func(t *testing.T) {
		// Create two secrets in the same directory
		secret1Path := "shared/secret1"
		secret2Path := "shared/secret2"

		fullPath1 := filepath.Join(tmpDir, secret1Path)
		fullPath2 := filepath.Join(tmpDir, secret2Path)

		err := os.MkdirAll(filepath.Dir(fullPath1), 0700)
		require.NoError(t, err)
		err = os.WriteFile(fullPath1, []byte("test1"), 0600)
		require.NoError(t, err)
		err = os.WriteFile(fullPath2, []byte("test2"), 0600)
		require.NoError(t, err)

		// Delete one secret
		err = provider.DeleteSecret(ctx, secret1Path)
		require.NoError(t, err)

		// Verify the directory still exists (because it contains secret2)
		sharedDir := filepath.Join(tmpDir, "shared")
		_, err = os.Stat(sharedDir)
		require.NoError(t, err)

		// Verify secret2 still exists
		_, err = provider.GetSecret(ctx, secret2Path)
		require.NoError(t, err)
	})
}

func TestProvider_ListSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	provider, err := NewProvider(tmpDir, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("lists all secrets with prefix", func(t *testing.T) {
		// Create several secrets
		secretPaths := []string{
			"app/prod/database",
			"app/prod/redis",
			"app/staging/database",
			"app/staging/redis",
			"other/secret",
		}

		for _, secretPath := range secretPaths {
			fullPath := filepath.Join(tmpDir, secretPath)
			err := os.MkdirAll(filepath.Dir(fullPath), 0700)
			require.NoError(t, err)
			err = os.WriteFile(fullPath, []byte("test"), 0600)
			require.NoError(t, err)
		}

		// List secrets with "app/prod" prefix
		paths, err := provider.ListSecrets(ctx, "app/prod")

		require.NoError(t, err)
		assert.Len(t, paths, 2)
		assert.Contains(t, paths, "app/prod/database")
		assert.Contains(t, paths, "app/prod/redis")
	})

	t.Run("returns empty list for non-existent prefix", func(t *testing.T) {
		paths, err := provider.ListSecrets(ctx, "non/existent")

		require.NoError(t, err)
		assert.Empty(t, paths)
	})

	t.Run("lists all secrets when prefix is empty", func(t *testing.T) {
		// Create a secret at root level
		secretPath := "root-secret"
		fullPath := filepath.Join(tmpDir, secretPath)
		err := os.WriteFile(fullPath, []byte("test"), 0600)
		require.NoError(t, err)

		// List all secrets
		paths, err := provider.ListSecrets(ctx, "")

		require.NoError(t, err)
		assert.NotEmpty(t, paths)
		assert.Contains(t, paths, "root-secret")
	})

	t.Run("includes all file types", func(t *testing.T) {
		// Create files with different extensions
		testFiles := []string{
			"secret.txt",
			"config.yml",
			"token", // no extension
			"key.pem",
		}

		for _, filename := range testFiles {
			fullPath := filepath.Join(tmpDir, filename)
			err := os.WriteFile(fullPath, []byte("content"), 0600)
			require.NoError(t, err)
		}

		// List secrets
		paths, err := provider.ListSecrets(ctx, "")

		require.NoError(t, err)
		// Should include all files regardless of extension
		for _, filename := range testFiles {
			assert.Contains(t, paths, filename)
		}
	})
}

func TestProvider_secretPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	provider, err := NewProvider(tmpDir, zap.NewNop())
	require.NoError(t, err)

	t.Run("converts logical path to file path", func(t *testing.T) {
		expected := filepath.Join(tmpDir, "app", "prod", "database")
		actual := provider.secretPath("app/prod/database")

		assert.Equal(t, expected, actual)
	})

	t.Run("sanitizes directory traversal attempts", func(t *testing.T) {
		// Test various directory traversal attempts
		testCases := []string{
			"../../../etc/passwd",
			"./../../etc/passwd",
			"app/../../../etc/passwd",
		}

		for _, testCase := range testCases {
			actual := provider.secretPath(testCase)

			// The path should be sanitized (no literal ".." should remain)
			assert.NotContains(t, actual, "..")
		}
	})

	t.Run("handles absolute paths", func(t *testing.T) {
		// Absolute paths should be converted to relative
		expected := filepath.Join(tmpDir, "app", "secret")
		actual := provider.secretPath("/app/secret")

		assert.Equal(t, expected, actual)
	})

	t.Run("no extension added", func(t *testing.T) {
		actual := provider.secretPath("test/secret")

		// Should not add any extension
		assert.Empty(t, filepath.Ext(actual))
		assert.Equal(t, "secret", filepath.Base(actual))
	})
}

func TestProvider_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	provider, err := NewProvider(tmpDir, zap.NewNop())
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("full lifecycle test", func(t *testing.T) {
		secretPath := "integration/test/secret"

		// Create secret
		secret := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				secretPath: []byte("secret123"),
			},
		}

		// Store the secret
		err := provider.PutSecret(ctx, secretPath, secret)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, secretPath)
		require.NoError(t, err)
		assert.Equal(t, secret.Data, retrieved.Data)

		// List secrets
		paths, err := provider.ListSecrets(ctx, "integration")
		require.NoError(t, err)
		assert.Contains(t, paths, secretPath)

		// Update secret
		secret.Data[secretPath] = []byte("new_password")

		err = provider.PutSecret(ctx, secretPath, secret)
		require.NoError(t, err)

		// Verify update
		updated, err := provider.GetSecret(ctx, secretPath)
		require.NoError(t, err)
		assert.Equal(t, []byte("new_password"), updated.Data[secretPath])

		// Delete secret
		err = provider.DeleteSecret(ctx, secretPath)
		require.NoError(t, err)

		// Verify deletion
		_, err = provider.GetSecret(ctx, secretPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found")
	})
}
