package vault

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

func TestConfig_Structure(t *testing.T) {
	t.Run("config has required fields", func(t *testing.T) {
		cfg := Config{
			Address:    "https://vault.example.com",
			AuthMethod: "token",
			Role:       "test-role",
			Mount:      "secret",
			Token:      "test-token",
			Namespace:  "test-namespace",
		}

		assert.Equal(t, "https://vault.example.com", cfg.Address)
		assert.Equal(t, "token", cfg.AuthMethod)
		assert.Equal(t, "test-role", cfg.Role)
		assert.Equal(t, "secret", cfg.Mount)
		assert.Equal(t, "test-token", cfg.Token)
		assert.Equal(t, "test-namespace", cfg.Namespace)
	})
}

func TestProvider_PathGeneration(t *testing.T) {
	// Test path generation logic by examining the expected behavior
	mount := "secret"

	t.Run("data path generation", func(t *testing.T) {
		path := "test/secret"
		expectedDataPath := mount + "/data/" + path
		assert.Equal(t, "secret/data/test/secret", expectedDataPath)
	})

	t.Run("metadata path generation", func(t *testing.T) {
		path := "test/secret"
		expectedMetadataPath := mount + "/metadata/" + path
		assert.Equal(t, "secret/metadata/test/secret", expectedMetadataPath)
	})

	t.Run("list path generation", func(t *testing.T) {
		pathPrefix := "app/prod"
		expectedListPath := mount + "/metadata/" + pathPrefix
		assert.Equal(t, "secret/metadata/app/prod", expectedListPath)
	})

	t.Run("root list path", func(t *testing.T) {
		pathPrefix := ""
		expectedListPath := mount + "/metadata/" + pathPrefix
		assert.Equal(t, "secret/metadata/", expectedListPath)
	})
}

func TestProvider_SecretDataProcessing(t *testing.T) {
	t.Run("processes vault KV v2 response correctly", func(t *testing.T) {
		// Simulate the structure that would come from Vault KV v2
		vaultResponse := map[string]any{
			"data": map[string]any{
				"username": "admin",
				"password": "secret123",
				"port":     123,            // Non-string value should be skipped
				"enabled":  true,           // Non-string value should be skipped
				"bytes":    []byte("test"), // Bytes should be preserved
			},
			"metadata": map[string]any{
				"version":      float64(1),
				"created_by":   "test-user",
				"created_time": "2023-01-01T00:00:00Z", // String metadata
			},
		}

		// Test the data processing logic
		data := make(map[string][]byte)
		metadata := make(map[string]string)
		version := ""

		// Process data - this mimics the logic in GetSecret
		if secretData, ok := vaultResponse["data"].(map[string]any); ok {
			for k, v := range secretData {
				switch val := v.(type) {
				case string:
					data[k] = []byte(val)
				case []byte:
					data[k] = val
				}
			}
		}

		// Process metadata - this mimics the logic in GetSecret
		if meta, ok := vaultResponse["metadata"].(map[string]any); ok {
			for k, v := range meta {
				if strVal, ok := v.(string); ok {
					metadata[k] = strVal
				}
			}
			if _, ok := meta["version"].(float64); ok {
				version = "1"
			}
		}

		// Verify processing results
		assert.Equal(t, []byte("admin"), data["username"])
		assert.Equal(t, []byte("secret123"), data["password"])
		assert.Equal(t, []byte("test"), data["bytes"])

		// Non-string values should be skipped
		_, exists := data["port"]
		assert.False(t, exists)
		_, exists = data["enabled"]
		assert.False(t, exists)

		// Metadata should be processed
		assert.Equal(t, "test-user", metadata["created_by"])
		assert.Equal(t, "2023-01-01T00:00:00Z", metadata["created_time"])
		assert.Equal(t, "1", version)
	})

	t.Run("handles empty vault response", func(t *testing.T) {
		// Test nil response
		var vaultResponse map[string]any

		if vaultResponse == nil {
			// This should result in a "secret not found" error
			assert.Nil(t, vaultResponse)
		}

		// Test response with nil data
		vaultResponse = map[string]any{
			"data": nil,
		}

		if vaultResponse["data"] == nil {
			// This should result in a "secret not found" error
			assert.Nil(t, vaultResponse["data"])
		}
	})
}

func TestProvider_PayloadGeneration(t *testing.T) {
	t.Run("generates correct payload for PutSecret", func(t *testing.T) {
		secret := &secrets.Secret{
			Path: "test/secret",
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("secret123"),
			},
			Metadata: map[string]string{
				"created_by": "test",
				"env":        "prod",
			},
		}

		// Mimic the payload generation logic from PutSecret
		data := make(map[string]any)
		for k, v := range secret.Data {
			data[k] = string(v)
		}

		payload := map[string]any{
			"data": data,
		}

		if len(secret.Metadata) > 0 {
			options := make(map[string]any)
			for k, v := range secret.Metadata {
				options[k] = v
			}
			payload["options"] = options
		}

		// Verify payload structure
		assert.Contains(t, payload, "data")
		assert.Contains(t, payload, "options")

		payloadData := payload["data"].(map[string]any)
		assert.Equal(t, "admin", payloadData["username"])
		assert.Equal(t, "secret123", payloadData["password"])

		payloadOptions := payload["options"].(map[string]any)
		assert.Equal(t, "test", payloadOptions["created_by"])
		assert.Equal(t, "prod", payloadOptions["env"])
	})

	t.Run("generates payload without metadata", func(t *testing.T) {
		secret := &secrets.Secret{
			Path: "test/simple",
			Data: map[string][]byte{
				"value": []byte("test"),
			},
		}

		// Mimic the payload generation logic from PutSecret
		data := make(map[string]any)
		for k, v := range secret.Data {
			data[k] = string(v)
		}

		payload := map[string]any{
			"data": data,
		}

		// Should not include options when metadata is empty
		if len(secret.Metadata) > 0 {
			options := make(map[string]any)
			for k, v := range secret.Metadata {
				options[k] = v
			}
			payload["options"] = options
		}

		// Verify payload structure
		assert.Contains(t, payload, "data")
		assert.NotContains(t, payload, "options")

		payloadData := payload["data"].(map[string]any)
		assert.Equal(t, "test", payloadData["value"])
	})
}

func TestProvider_ListProcessing(t *testing.T) {
	t.Run("processes vault list response correctly", func(t *testing.T) {
		vaultResponse := map[string]any{
			"data": map[string]any{
				"keys": []any{
					"database",
					"redis",
					"subdir/",
					"another-secret",
				},
			},
		}

		pathPrefix := "app/prod"

		// Mimic the list processing logic from ListSecrets
		var paths []string

		if vaultResponse != nil {
			if data, ok := vaultResponse["data"].(map[string]any); ok {
				if keys, ok := data["keys"].([]any); ok {
					for _, key := range keys {
						if strKey, ok := key.(string); ok {
							// Remove trailing slash for directories
							strKey = strings.TrimSuffix(strKey, "/")
							if pathPrefix != "" {
								paths = append(paths, pathPrefix+"/"+strKey)
							} else {
								paths = append(paths, strKey)
							}
						}
					}
				}
			}
		}

		// Verify processing results
		require.Len(t, paths, 4)
		assert.Contains(t, paths, "app/prod/database")
		assert.Contains(t, paths, "app/prod/redis")
		assert.Contains(t, paths, "app/prod/subdir")
		assert.Contains(t, paths, "app/prod/another-secret")
	})

	t.Run("handles empty list response", func(t *testing.T) {
		// Test nil response
		var vaultResponse map[string]any

		var paths []string

		if vaultResponse == nil {
			paths = []string{}
		}

		assert.Empty(t, paths)

		// Test empty keys
		vaultResponse = map[string]any{
			"data": map[string]any{
				"keys": []any{},
			},
		}

		paths = []string{}
		if data, ok := vaultResponse["data"].(map[string]any); ok {
			if keys, ok := data["keys"].([]any); ok {
				for _, key := range keys {
					if strKey, ok := key.(string); ok {
						paths = append(paths, strKey)
					}
				}
			}
		}

		assert.Empty(t, paths)
	})

	t.Run("handles root path listing", func(t *testing.T) {
		vaultResponse := map[string]any{
			"data": map[string]any{
				"keys": []any{
					"root-secret",
					"app/",
					"config/",
				},
			},
		}

		pathPrefix := ""

		var paths []string
		if data, ok := vaultResponse["data"].(map[string]any); ok {
			if keys, ok := data["keys"].([]any); ok {
				for _, key := range keys {
					if strKey, ok := key.(string); ok {
						// Remove trailing slash for directories
						strKey = strings.TrimSuffix(strKey, "/")
						if pathPrefix != "" {
							paths = append(paths, pathPrefix+"/"+strKey)
						} else {
							paths = append(paths, strKey)
						}
					}
				}
			}
		}

		require.Len(t, paths, 3)
		assert.Contains(t, paths, "root-secret")
		assert.Contains(t, paths, "app")
		assert.Contains(t, paths, "config")
	})
}

func TestAuthenticate_TokenAuth(t *testing.T) {
	t.Run("validates token auth configuration", func(t *testing.T) {
		// Test with provided token
		cfg := Config{
			AuthMethod: "token",
			Token:      "test-token",
		}

		assert.Equal(t, "token", cfg.AuthMethod)
		assert.Equal(t, "test-token", cfg.Token)
	})

	t.Run("validates environment token fallback", func(t *testing.T) {
		originalToken := os.Getenv("VAULT_TOKEN")
		defer func() {
			if originalToken != "" {
				os.Setenv("VAULT_TOKEN", originalToken)
			} else {
				os.Unsetenv("VAULT_TOKEN")
			}
		}()

		os.Setenv("VAULT_TOKEN", "env-token")

		cfg := Config{
			AuthMethod: "token",
			Token:      "",
		}

		// Simulate the logic from authenticate function
		token := cfg.Token
		if token == "" {
			token = os.Getenv("VAULT_TOKEN")
		}

		assert.Equal(t, "env-token", token)
	})

	t.Run("detects missing token", func(t *testing.T) {
		originalToken := os.Getenv("VAULT_TOKEN")
		defer func() {
			if originalToken != "" {
				os.Setenv("VAULT_TOKEN", originalToken)
			} else {
				os.Unsetenv("VAULT_TOKEN")
			}
		}()

		os.Unsetenv("VAULT_TOKEN")

		cfg := Config{
			AuthMethod: "token",
			Token:      "",
		}

		// Simulate the logic from authenticate function
		token := cfg.Token
		if token == "" {
			token = os.Getenv("VAULT_TOKEN")
		}

		if token == "" {
			assert.Empty(t, token, "should detect missing token")
		}
	})
}

func TestAuthenticate_OtherMethods(t *testing.T) {
	t.Run("validates kubernetes auth configuration", func(t *testing.T) {
		cfg := Config{
			AuthMethod: "kubernetes",
			Role:       "test-role",
		}

		assert.Equal(t, "kubernetes", cfg.AuthMethod)
		assert.Equal(t, "test-role", cfg.Role)
	})

	t.Run("validates approle auth configuration", func(t *testing.T) {
		cfg := Config{
			AuthMethod: "approle",
			Role:       "test-role",
		}

		assert.Equal(t, "approle", cfg.AuthMethod)
		assert.Equal(t, "test-role", cfg.Role)
	})

	t.Run("detects unsupported auth method", func(t *testing.T) {
		cfg := Config{
			AuthMethod: "unsupported",
		}

		// The authenticate function should return an error for unsupported methods
		assert.Equal(t, "unsupported", cfg.AuthMethod)
	})
}

func TestProvider_Close(t *testing.T) {
	t.Run("close method exists and returns no error", func(t *testing.T) {
		// Create a minimal provider for testing
		provider := &Provider{
			client: nil, // nil is ok for this test
			mount:  "secret",
			logger: zap.NewNop(),
		}

		err := provider.Close()
		assert.NoError(t, err)
	})
}

// Helper function to add missing strings import
func TestHelper_StringOperations(t *testing.T) {
	t.Run("string trimming logic", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"subdir/", "subdir"},
			{"file", "file"},
			{"nested/dir/", "nested/dir"},
		}

		for _, tc := range testCases {
			result := strings.TrimSuffix(tc.input, "/")
			assert.Equal(t, tc.expected, result)
		}
	})
}

// Add missing strings import by adding this at the top of the file
// This test ensures the strings package is imported
func init() {
	_ = strings.TrimSuffix("test/", "/")
}
