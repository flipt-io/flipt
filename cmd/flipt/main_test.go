package main

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

// mockManager is a test mock for secrets.Manager
type mockManager struct {
	secrets map[string][]byte
}

func newMockManager(secrets map[string][]byte) *mockManager {
	return &mockManager{secrets: secrets}
}

func (m *mockManager) RegisterProvider(name string, provider secrets.Provider) error {
	return nil
}

func (m *mockManager) GetProvider(name string) (secrets.Provider, error) {
	return nil, nil
}

func (m *mockManager) GetSecretValue(ctx context.Context, ref secrets.Reference) ([]byte, error) {
	key := ref.Provider + ":" + ref.Key
	if value, ok := m.secrets[key]; ok {
		return value, nil
	}
	return nil, nil
}

func (m *mockManager) GetSecret(ctx context.Context, providerName, path string) (*secrets.Secret, error) {
	return nil, nil
}

func (m *mockManager) ListSecrets(ctx context.Context, providerName, pathPrefix string) ([]string, error) {
	return nil, nil
}

func (m *mockManager) ListProviders() []string {
	return nil
}

func (m *mockManager) Close() error {
	return nil
}

// testConfig is a simple struct for testing
type testConfig struct {
	SimpleField string
	NestedField struct {
		Value string
	}
	MapField map[string]testMapValue
}

type testMapValue struct {
	ID     string
	Secret string
}

func TestWalkConfigForSecrets_SimpleField(t *testing.T) {
	cfg := &testConfig{
		SimpleField: "${secret:file:mykey}",
	}

	manager := newMockManager(map[string][]byte{
		"file:mykey": []byte("resolved-value"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)
	assert.Equal(t, "resolved-value", cfg.SimpleField)
}

func TestWalkConfigForSecrets_NestedField(t *testing.T) {
	cfg := &testConfig{
		SimpleField: "plain-value",
		NestedField: struct{ Value string }{
			Value: "${secret:file:nested}",
		},
	}

	manager := newMockManager(map[string][]byte{
		"file:nested": []byte("nested-resolved"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)
	assert.Equal(t, "plain-value", cfg.SimpleField)
	assert.Equal(t, "nested-resolved", cfg.NestedField.Value)
}

func TestWalkConfigForSecrets_MapWithStructValues(t *testing.T) {
	// This test verifies the fix for the issue where secret references
	// in map values (like OIDC provider credentials) were not being resolved
	// because map values from MapIndex() are not addressable.
	cfg := &testConfig{
		MapField: map[string]testMapValue{
			"provider1": {
				ID:     "${secret:file:clientid}",
				Secret: "${secret:file:clientsecret}",
			},
		},
	}

	manager := newMockManager(map[string][]byte{
		"file:clientid":     []byte("my-client-id"),
		"file:clientsecret": []byte("my-client-secret"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	// Verify the map values were updated with resolved secrets
	provider := cfg.MapField["provider1"]
	assert.Equal(t, "my-client-id", provider.ID)
	assert.Equal(t, "my-client-secret", provider.Secret)
}

func TestWalkConfigForSecrets_MapWithMultipleProviders(t *testing.T) {
	cfg := &testConfig{
		MapField: map[string]testMapValue{
			"keycloak": {
				ID:     "${secret:file:keycloak-id}",
				Secret: "${secret:file:keycloak-secret}",
			},
			"google": {
				ID:     "${secret:file:google-id}",
				Secret: "${secret:file:google-secret}",
			},
		},
	}

	manager := newMockManager(map[string][]byte{
		"file:keycloak-id":     []byte("keycloak-client-id"),
		"file:keycloak-secret": []byte("keycloak-client-secret"),
		"file:google-id":       []byte("google-client-id"),
		"file:google-secret":   []byte("google-client-secret"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	keycloak := cfg.MapField["keycloak"]
	assert.Equal(t, "keycloak-client-id", keycloak.ID)
	assert.Equal(t, "keycloak-client-secret", keycloak.Secret)

	google := cfg.MapField["google"]
	assert.Equal(t, "google-client-id", google.ID)
	assert.Equal(t, "google-client-secret", google.Secret)
}

func TestWalkConfigForSecrets_GCPProvider(t *testing.T) {
	cfg := &testConfig{
		SimpleField: "${secret:gcp:my-api-key}",
	}

	manager := newMockManager(map[string][]byte{
		"gcp:my-api-key": []byte("gcp-resolved-value"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)
	assert.Equal(t, "gcp-resolved-value", cfg.SimpleField)
}

func TestWalkConfigForSecrets_GCPProviderInMap(t *testing.T) {
	cfg := &testConfig{
		MapField: map[string]testMapValue{
			"provider1": {
				ID:     "${secret:gcp:client-id}",
				Secret: "${secret:gcp:client-secret}",
			},
		},
	}

	manager := newMockManager(map[string][]byte{
		"gcp:client-id":     []byte("gcp-client-id"),
		"gcp:client-secret": []byte("gcp-client-secret"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	provider := cfg.MapField["provider1"]
	assert.Equal(t, "gcp-client-id", provider.ID)
	assert.Equal(t, "gcp-client-secret", provider.Secret)
}

func TestWalkConfigForSecrets_AWSProvider(t *testing.T) {
	cfg := &testConfig{
		SimpleField: "${secret:aws:my-api-key}",
	}

	manager := newMockManager(map[string][]byte{
		"aws:my-api-key": []byte("aws-resolved-value"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)
	assert.Equal(t, "aws-resolved-value", cfg.SimpleField)
}

func TestWalkConfigForSecrets_AWSProviderInMap(t *testing.T) {
	cfg := &testConfig{
		MapField: map[string]testMapValue{
			"provider1": {
				ID:     "${secret:aws:client-id}",
				Secret: "${secret:aws:client-secret}",
			},
		},
	}

	manager := newMockManager(map[string][]byte{
		"aws:client-id":     []byte("aws-client-id"),
		"aws:client-secret": []byte("aws-client-secret"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	provider := cfg.MapField["provider1"]
	assert.Equal(t, "aws-client-id", provider.ID)
	assert.Equal(t, "aws-client-secret", provider.Secret)
}

func TestWalkConfigForSecrets_MixedProviders(t *testing.T) {
	cfg := &testConfig{
		SimpleField: "${secret:vault:db-password}",
		MapField: map[string]testMapValue{
			"keycloak": {
				ID:     "${secret:file:keycloak-id}",
				Secret: "${secret:gcp:keycloak-secret}",
			},
			"aws-service": {
				ID:     "${secret:aws:service-id}",
				Secret: "${secret:aws:service-secret}",
			},
			"azure-service": {
				ID:     "${secret:azure:azure-id}",
				Secret: "${secret:azure:azure-secret}",
			},
		},
	}

	manager := newMockManager(map[string][]byte{
		"vault:db-password":   []byte("vault-db-pass"),
		"file:keycloak-id":    []byte("file-keycloak-id"),
		"gcp:keycloak-secret": []byte("gcp-keycloak-secret"),
		"aws:service-id":      []byte("aws-service-id"),
		"aws:service-secret":  []byte("aws-service-secret"),
		"azure:azure-id":      []byte("azure-service-id"),
		"azure:azure-secret":  []byte("azure-service-secret"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	assert.Equal(t, "vault-db-pass", cfg.SimpleField)

	keycloak := cfg.MapField["keycloak"]
	assert.Equal(t, "file-keycloak-id", keycloak.ID)
	assert.Equal(t, "gcp-keycloak-secret", keycloak.Secret)

	awsService := cfg.MapField["aws-service"]
	assert.Equal(t, "aws-service-id", awsService.ID)
	assert.Equal(t, "aws-service-secret", awsService.Secret)

	azureService := cfg.MapField["azure-service"]
	assert.Equal(t, "azure-service-id", azureService.ID)
	assert.Equal(t, "azure-service-secret", azureService.Secret)
}

func TestWalkConfigForSecrets_NoSecretReferences(t *testing.T) {
	cfg := &testConfig{
		SimpleField: "plain-value",
		MapField: map[string]testMapValue{
			"provider1": {
				ID:     "hardcoded-id",
				Secret: "hardcoded-secret",
			},
		},
	}

	manager := newMockManager(map[string][]byte{})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	// Values should remain unchanged
	assert.Equal(t, "plain-value", cfg.SimpleField)
	provider := cfg.MapField["provider1"]
	assert.Equal(t, "hardcoded-id", provider.ID)
	assert.Equal(t, "hardcoded-secret", provider.Secret)
}

func TestWalkConfigForSecrets_MixedValues(t *testing.T) {
	// Mix of secret references and plain values
	cfg := &testConfig{
		SimpleField: "plain-value",
		MapField: map[string]testMapValue{
			"provider1": {
				ID:     "${secret:file:clientid}",
				Secret: "hardcoded-secret",
			},
		},
	}

	manager := newMockManager(map[string][]byte{
		"file:clientid": []byte("resolved-client-id"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	provider := cfg.MapField["provider1"]
	assert.Equal(t, "resolved-client-id", provider.ID)
	assert.Equal(t, "hardcoded-secret", provider.Secret)
}

func TestWalkConfigForSecrets_AzureProvider(t *testing.T) {
	cfg := &testConfig{
		SimpleField: "${secret:azure:my-api-key}",
	}

	manager := newMockManager(map[string][]byte{
		"azure:my-api-key": []byte("azure-resolved-value"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)
	assert.Equal(t, "azure-resolved-value", cfg.SimpleField)
}

func TestWalkConfigForSecrets_AzureProviderInMap(t *testing.T) {
	cfg := &testConfig{
		MapField: map[string]testMapValue{
			"provider1": {
				ID:     "${secret:azure:client-id}",
				Secret: "${secret:azure:client-secret}",
			},
		},
	}

	manager := newMockManager(map[string][]byte{
		"azure:client-id":     []byte("azure-client-id"),
		"azure:client-secret": []byte("azure-client-secret"),
	})

	err := walkConfigForSecrets(t.Context(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	provider := cfg.MapField["provider1"]
	assert.Equal(t, "azure-client-id", provider.ID)
	assert.Equal(t, "azure-client-secret", provider.Secret)
}

func TestResolveLicenseSecrets(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "license-key"), []byte("file-license-key\n"), 0o600))

	t.Run("resolves file provider references", func(t *testing.T) {
		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				Providers: config.ProvidersConfig{
					File: &config.FileProviderConfig{Enabled: true, BasePath: dir},
				},
			},
			License: config.LicenseConfig{Key: "${secret:file:license-key}"},
		}
		cfg.Meta.StateDirectory = "${secret:file:license-key}"

		require.NoError(t, resolveLicenseSecrets(t.Context(), zap.NewNop(), cfg))
		assert.Equal(t, "file-license-key", cfg.License.Key)
		// only license.* is resolved before the license is known
		assert.Equal(t, "${secret:file:license-key}", cfg.Meta.StateDirectory)
	})

	t.Run("does not use pro providers", func(t *testing.T) {
		original, ok := secrets.GetProviderFactory("vault")
		require.True(t, ok)

		called := false
		secrets.RegisterProviderFactory("vault", func(*config.Config, *zap.Logger) (secrets.Provider, error) {
			called = true
			return nil, nil
		})
		t.Cleanup(func() { secrets.RegisterProviderFactory("vault", original) })

		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				Providers: config.ProvidersConfig{
					File:  &config.FileProviderConfig{Enabled: true, BasePath: dir},
					Vault: &config.VaultProviderConfig{Enabled: true},
				},
			},
			License: config.LicenseConfig{Key: "${secret:vault:license-key}"},
		}

		err := resolveLicenseSecrets(t.Context(), zap.NewNop(), cfg)
		require.EqualError(t, err, `license.key: secret reference "${secret:vault:license-key}" must use the file provider`)
		assert.Equal(t, "${secret:vault:license-key}", cfg.License.Key)
		// resolveLicenseSecrets only builds the file provider, so this guards
		// against a regression that sets up providers from the full config
		assert.False(t, called, "vault factory must not be invoked to resolve the license")
	})

	t.Run("rejects references without a provider", func(t *testing.T) {
		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				Providers: config.ProvidersConfig{
					File: &config.FileProviderConfig{Enabled: true, BasePath: dir},
				},
			},
			License: config.LicenseConfig{MachineID: "${secret:license-key}"},
		}

		err := resolveLicenseSecrets(t.Context(), zap.NewNop(), cfg)
		require.EqualError(t, err, `license.machine_id: secret reference "${secret:license-key}" must use the file provider`)
	})

	t.Run("requires the file provider to be enabled", func(t *testing.T) {
		cfg := &config.Config{
			License: config.LicenseConfig{Key: "${secret:file:license-key}"},
		}

		err := resolveLicenseSecrets(t.Context(), zap.NewNop(), cfg)
		require.EqualError(t, err, `license.key: secret reference "${secret:file:license-key}" requires the file secrets provider to be enabled (secrets.providers.file.enabled)`)
	})
}

func TestResolveSecretsInConfig_SkipsLicense(t *testing.T) {
	// a license value that was already resolved from the file provider must
	// not be resolved again, even if its content looks like a reference
	cfg := &config.Config{
		License: config.LicenseConfig{Key: "${secret:vault:license-key}"},
	}
	cfg.Meta.StateDirectory = "${secret:vault:state-dir}"

	manager := newMockManager(map[string][]byte{
		"vault:license-key": []byte("from-vault"),
		"vault:state-dir":   []byte("/var/lib/flipt"),
	})

	require.NoError(t, resolveSecretsInConfig(t.Context(), cfg, manager))
	assert.Equal(t, "${secret:vault:license-key}", cfg.License.Key)
	assert.Equal(t, "/var/lib/flipt", cfg.Meta.StateDirectory)
}
