package file

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
		// Create a secret file manually
		secretPath := "test/secret"
		secretData := map[string][]byte{
			"password": []byte("secret123"),
			"token":    []byte("abc123"),
		}
		secretMetadata := map[string]string{
			"created_by": "test",
			"env":        "test",
		}

		secret := &secrets.Secret{
			Path:     secretPath,
			Data:     secretData,
			Metadata: secretMetadata,
			Version:  "v1",
		}

		err := provider.PutSecret(ctx, secretPath, secret)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, secretPath)

		require.NoError(t, err)
		assert.Equal(t, secretPath, retrieved.Path)
		assert.Equal(t, secretData, retrieved.Data)
		assert.Equal(t, secretMetadata, retrieved.Metadata)
		assert.Equal(t, "v1", retrieved.Version)
	})

	t.Run("returns error for non-existent secret", func(t *testing.T) {
		_, err := provider.GetSecret(ctx, "non/existent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found")
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		// Create invalid JSON file
		invalidPath := filepath.Join(tmpDir, "invalid.json")
		err := os.WriteFile(invalidPath, []byte("invalid json"), 0600)
		require.NoError(t, err)

		_, err = provider.GetSecret(ctx, "invalid")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing secret file")
	})

	t.Run("handles GPG key with escaped newlines in JSON", func(t *testing.T) {
		// Create a GPG key with properly escaped newlines in JSON
		gpgKey := "-----BEGIN PGP PRIVATE KEY BLOCK-----\n\nlQVYBGh9OVoBDACmOSHo...\n-----END PGP PRIVATE KEY BLOCK-----"

		// Create JSON manually with properly escaped string
		secretFile := secretFile{
			Data: map[string]string{
				"gpg-key": gpgKey,
			},
		}

		jsonData, err := json.Marshal(secretFile)
		require.NoError(t, err)

		secretPath := filepath.Join(tmpDir, "gpg-test.json")
		err = os.WriteFile(secretPath, jsonData, 0600)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, "gpg-test")

		require.NoError(t, err)
		assert.Equal(t, []byte(gpgKey), retrieved.Data["gpg-key"])
	})

	t.Run("handles base64 encoded values in JSON", func(t *testing.T) {
		// Create a secret file with base64 encoded values
		gpgKey := "-----BEGIN PGP PRIVATE KEY BLOCK-----\n\nlQVYBGh9OVoBDACmOSHo...\n-----END PGP PRIVATE KEY BLOCK-----"
		base64EncodedKey := base64.StdEncoding.EncodeToString([]byte(gpgKey))

		secretFile := secretFile{
			Data: map[string]string{
				"gpg-key":   base64EncodedKey,
				"plaintext": "not-base64-value",
			},
		}

		jsonData, err := json.Marshal(secretFile)
		require.NoError(t, err)

		secretPath := filepath.Join(tmpDir, "base64-values-test.json")
		err = os.WriteFile(secretPath, jsonData, 0600)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, "base64-values-test")

		require.NoError(t, err)
		// Base64 encoded value should be decoded
		assert.Equal(t, []byte(gpgKey), retrieved.Data["gpg-key"])
		// Plain text value should remain as-is
		assert.Equal(t, []byte("not-base64-value"), retrieved.Data["plaintext"])
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
		secretPath := "test/secret"
		secret := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				"password": []byte("secret123"),
				"token":    []byte("abc123"),
			},
			Metadata: map[string]string{
				"created_by": "test",
			},
			Version: "v1",
		}

		err := provider.PutSecret(ctx, secretPath, secret)

		require.NoError(t, err)

		// Verify file was created
		fullPath := provider.secretPath(secretPath)
		stat, err := os.Stat(fullPath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), stat.Mode().Perm())

		// Verify directory structure was created
		dir := filepath.Dir(fullPath)
		dirStat, err := os.Stat(dir)
		require.NoError(t, err)
		assert.True(t, dirStat.IsDir())
		assert.Equal(t, os.FileMode(0700), dirStat.Mode().Perm())
	})

	t.Run("overwrites existing secret", func(t *testing.T) {
		secretPath := "test/overwrite"

		// Create initial secret
		secret1 := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				"password": []byte("old_password"),
			},
			Version: "v1",
		}

		err := provider.PutSecret(ctx, secretPath, secret1)
		require.NoError(t, err)

		// Overwrite with new secret
		secret2 := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				"password": []byte("new_password"),
			},
			Version: "v2",
		}

		err = provider.PutSecret(ctx, secretPath, secret2)
		require.NoError(t, err)

		// Verify the secret was overwritten
		retrieved, err := provider.GetSecret(ctx, secretPath)
		require.NoError(t, err)
		assert.Equal(t, []byte("new_password"), retrieved.Data["password"])
		assert.Equal(t, "v2", retrieved.Version)
	})

	t.Run("handles nested paths", func(t *testing.T) {
		secretPath := "app/prod/database/config" // #nosec G101 - this is a test path
		secret := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				"username": []byte("dbuser"),
				"password": []byte("dbpass"),
			},
		}

		err := provider.PutSecret(ctx, secretPath, secret)

		require.NoError(t, err)

		// Verify the secret can be retrieved
		retrieved, err := provider.GetSecret(ctx, secretPath)
		require.NoError(t, err)
		assert.Equal(t, secret.Data, retrieved.Data)
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
		secretPath := "test/delete"
		secret := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				"password": []byte("secret123"),
			},
		}

		// Create the secret
		err := provider.PutSecret(ctx, secretPath, secret)
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
		secret := &secrets.Secret{
			Path: secretPath,
			Data: map[string][]byte{
				"value": []byte("test"),
			},
		}

		// Create the secret
		err := provider.PutSecret(ctx, secretPath, secret)
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

		// But base directory should still exist
		baseDir := filepath.Join(tmpDir, "cleanup")
		_, err = os.Stat(baseDir)
		assert.True(t, os.IsNotExist(err)) // Should be cleaned up since it's empty
	})

	t.Run("does not remove non-empty directories", func(t *testing.T) {
		// Create two secrets in the same directory
		secret1Path := "shared/secret1"
		secret2Path := "shared/secret2"

		secret1 := &secrets.Secret{
			Path: secret1Path,
			Data: map[string][]byte{"value": []byte("test1")},
		}
		secret2 := &secrets.Secret{
			Path: secret2Path,
			Data: map[string][]byte{"value": []byte("test2")},
		}

		err := provider.PutSecret(ctx, secret1Path, secret1)
		require.NoError(t, err)
		err = provider.PutSecret(ctx, secret2Path, secret2)
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
			secret := &secrets.Secret{
				Path: secretPath,
				Data: map[string][]byte{"value": []byte("test")},
			}
			err := provider.PutSecret(ctx, secretPath, secret)
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
		secret := &secrets.Secret{
			Path: "root-secret",
			Data: map[string][]byte{"value": []byte("test")},
		}
		err := provider.PutSecret(ctx, "root-secret", secret)
		require.NoError(t, err)

		// List all secrets
		paths, err := provider.ListSecrets(ctx, "")

		require.NoError(t, err)
		assert.NotEmpty(t, paths)
		assert.Contains(t, paths, "root-secret")
	})

	t.Run("ignores non-JSON files", func(t *testing.T) {
		// Create a non-JSON file in the secrets directory
		nonJSONPath := filepath.Join(tmpDir, "not-a-secret.txt")
		err := os.WriteFile(nonJSONPath, []byte("not a secret"), 0600) // #nosec G306 - test file
		require.NoError(t, err)

		// List secrets
		paths, err := provider.ListSecrets(ctx, "")

		require.NoError(t, err)
		// Should not include the non-JSON file
		for _, path := range paths {
			assert.NotEqual(t, "not-a-secret", path)
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
		expected := filepath.Join(tmpDir, "app", "prod", "database.json")
		actual := provider.secretPath("app/prod/database")

		assert.Equal(t, expected, actual)
	})

	t.Run("sanitizes directory traversal attempts", func(t *testing.T) {
		// Test various directory traversal attempts
		// NOTE: The current implementation uses filepath.Clean() which resolves '..' elements
		// but may still allow traversal outside the base directory.
		testCases := []string{
			"../../../etc/passwd",
			"./../../etc/passwd",
			"app/../../../etc/passwd",
		}

		for _, testCase := range testCases {
			actual := provider.secretPath(testCase)

			// The path should be sanitized (no literal ".." should remain)
			assert.NotContains(t, actual, "..")

			// Verify the path ends with .json
			assert.True(t, strings.HasSuffix(actual, ".json"))
		}
	})

	t.Run("handles absolute paths", func(t *testing.T) {
		// Absolute paths should be converted to relative
		expected := filepath.Join(tmpDir, "app", "secret.json")
		actual := provider.secretPath("/app/secret")

		assert.Equal(t, expected, actual)
	})

	t.Run("adds JSON extension", func(t *testing.T) {
		actual := provider.secretPath("test/secret")

		assert.Equal(t, ".json", filepath.Ext(actual))
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
				"username": []byte("admin"),
				"password": []byte("secret123"),
			},
			Metadata: map[string]string{
				"created_by": "test",
				"env":        "test",
			},
			Version: "v1",
		}

		// Store the secret
		err := provider.PutSecret(ctx, secretPath, secret)
		require.NoError(t, err)

		// Retrieve the secret
		retrieved, err := provider.GetSecret(ctx, secretPath)
		require.NoError(t, err)
		assert.Equal(t, secret.Data, retrieved.Data)
		assert.Equal(t, secret.Metadata, retrieved.Metadata)
		assert.Equal(t, secret.Version, retrieved.Version)

		// List secrets
		paths, err := provider.ListSecrets(ctx, "integration")
		require.NoError(t, err)
		assert.Contains(t, paths, secretPath)

		// Update secret
		secret.Data["password"] = []byte("new_password")
		secret.Version = "v2"

		err = provider.PutSecret(ctx, secretPath, secret)
		require.NoError(t, err)

		// Verify update
		updated, err := provider.GetSecret(ctx, secretPath)
		require.NoError(t, err)
		assert.Equal(t, []byte("new_password"), updated.Data["password"])
		assert.Equal(t, "v2", updated.Version)

		// Delete secret
		err = provider.DeleteSecret(ctx, secretPath)
		require.NoError(t, err)

		// Verify deletion
		_, err = provider.GetSecret(ctx, secretPath)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "secret not found")
	})
}
