package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/secrets"
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

	err := walkConfigForSecrets(context.Background(), reflect.ValueOf(cfg).Elem(), manager)
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

	err := walkConfigForSecrets(context.Background(), reflect.ValueOf(cfg).Elem(), manager)
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

	err := walkConfigForSecrets(context.Background(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	keycloak := cfg.MapField["keycloak"]
	assert.Equal(t, "keycloak-client-id", keycloak.ID)
	assert.Equal(t, "keycloak-client-secret", keycloak.Secret)

	google := cfg.MapField["google"]
	assert.Equal(t, "google-client-id", google.ID)
	assert.Equal(t, "google-client-secret", google.Secret)
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

	err := walkConfigForSecrets(context.Background(), reflect.ValueOf(cfg).Elem(), manager)
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

	err := walkConfigForSecrets(context.Background(), reflect.ValueOf(cfg).Elem(), manager)
	require.NoError(t, err)

	provider := cfg.MapField["provider1"]
	assert.Equal(t, "resolved-client-id", provider.ID)
	assert.Equal(t, "hardcoded-secret", provider.Secret)
}
