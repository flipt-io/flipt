package secrets

import (
	"context"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
)

func TestRegisterProviderFactory(t *testing.T) {
	// Clear the global registry before testing
	factoryMu.Lock()
	originalFactories := make(map[string]ProviderFactory)
	maps.Copy(originalFactories, providerFactories)
	providerFactories = make(map[string]ProviderFactory)
	factoryMu.Unlock()

	// Restore original factories after test
	t.Cleanup(func() {
		factoryMu.Lock()
		providerFactories = originalFactories
		factoryMu.Unlock()
	})

	t.Run("registers provider factory", func(t *testing.T) {
		factory := func(cfg *config.Config, logger *zap.Logger) (Provider, error) {
			return &MockProvider{}, nil
		}

		RegisterProviderFactory("test", factory)

		factoryMu.RLock()
		registeredFactory, exists := providerFactories["test"]
		factoryMu.RUnlock()

		assert.True(t, exists)
		assert.NotNil(t, registeredFactory)
	})

	t.Run("overwrites existing factory", func(t *testing.T) {
		factory1 := func(cfg *config.Config, logger *zap.Logger) (Provider, error) {
			return &MockProvider{}, nil
		}

		factory2 := func(cfg *config.Config, logger *zap.Logger) (Provider, error) {
			return &MockProvider{}, nil
		}

		RegisterProviderFactory("overwrite", factory1)
		RegisterProviderFactory("overwrite", factory2)

		factoryMu.RLock()
		registeredFactory, exists := providerFactories["overwrite"]
		factoryMu.RUnlock()

		assert.True(t, exists)
		assert.NotNil(t, registeredFactory)
		// We can't easily compare function pointers, but we verified it exists
	})
}

func TestNewManager(t *testing.T) {
	logger := zap.NewNop()

	t.Run("creates manager with file provider", func(t *testing.T) {
		// Clear and setup test factory
		factoryMu.Lock()
		originalFactories := make(map[string]ProviderFactory)
		maps.Copy(originalFactories, providerFactories)

		providerFactories = map[string]ProviderFactory{
			"file": func(cfg *config.Config, logger *zap.Logger) (Provider, error) {
				return &MockProvider{}, nil
			},
		}
		factoryMu.Unlock()

		t.Cleanup(func() {
			factoryMu.Lock()
			providerFactories = originalFactories
			factoryMu.Unlock()
		})

		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				Providers: config.ProvidersConfig{
					File: &config.FileProviderConfig{
						Enabled:  true,
						BasePath: "/tmp/secrets",
					},
				},
			},
		}

		manager, err := NewManager(logger, cfg)

		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Verify file provider was registered
		providers := manager.ListProviders()
		assert.Contains(t, providers, "file")
	})

	t.Run("creates manager with vault provider", func(t *testing.T) {
		// Clear and setup test factories
		factoryMu.Lock()
		originalFactories := make(map[string]ProviderFactory)
		maps.Copy(originalFactories, providerFactories)

		providerFactories = map[string]ProviderFactory{
			"file": func(cfg *config.Config, logger *zap.Logger) (Provider, error) {
				return &MockProvider{}, nil
			},
			"vault": func(cfg *config.Config, logger *zap.Logger) (Provider, error) {
				return &MockProvider{}, nil
			},
		}
		factoryMu.Unlock()

		t.Cleanup(func() {
			factoryMu.Lock()
			providerFactories = originalFactories
			factoryMu.Unlock()
		})

		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				Providers: config.ProvidersConfig{
					File: &config.FileProviderConfig{
						Enabled:  true,
						BasePath: "/tmp/secrets",
					},
					Vault: &config.VaultProviderConfig{
						Enabled:    true,
						Address:    "https://vault.example.com",
						AuthMethod: "token",
						Mount:      "secret",
						Token:      "test-token",
					},
				},
			},
		}

		manager, err := NewManager(logger, cfg)

		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Verify both providers were registered
		providers := manager.ListProviders()
		assert.Contains(t, providers, "file")
		assert.Contains(t, providers, "vault")
	})

	t.Run("fails when file factory not registered", func(t *testing.T) {
		// Clear factories
		factoryMu.Lock()
		originalFactories := make(map[string]ProviderFactory)
		maps.Copy(originalFactories, providerFactories)
		providerFactories = make(map[string]ProviderFactory)
		factoryMu.Unlock()

		t.Cleanup(func() {
			factoryMu.Lock()
			providerFactories = originalFactories
			factoryMu.Unlock()
		})

		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				Providers: config.ProvidersConfig{
					File: &config.FileProviderConfig{
						Enabled:  true,
						BasePath: "/tmp/secrets",
					},
				},
			},
		}

		_, err := NewManager(logger, cfg)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "file provider factory not registered")
	})

	t.Run("fails when Vault factory not registered", func(t *testing.T) {
		// Clear factories
		factoryMu.Lock()
		originalFactories := make(map[string]ProviderFactory)
		maps.Copy(originalFactories, providerFactories)
		providerFactories = make(map[string]ProviderFactory)
		factoryMu.Unlock()

		t.Cleanup(func() {
			factoryMu.Lock()
			providerFactories = originalFactories
			factoryMu.Unlock()
		})

		cfg := &config.Config{
			Secrets: config.SecretsConfig{
				Providers: config.ProvidersConfig{
					Vault: &config.VaultProviderConfig{
						Enabled:    true,
						Address:    "https://vault.example.com",
						AuthMethod: "token",
						Mount:      "secret",
						Token:      "test-token",
					},
				},
			},
		}

		_, err := NewManager(logger, cfg)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "vault provider factory not registered")
	})
}

func TestManagerImpl_RegisterProvider(t *testing.T) {
	manager := &ManagerImpl{
		providers: make(map[string]Provider),
		logger:    zap.NewNop(),
	}

	t.Run("registers provider successfully", func(t *testing.T) {
		provider := &MockProvider{}

		err := manager.RegisterProvider("test", provider)

		require.NoError(t, err)

		// Verify provider was registered
		registeredProvider, err := manager.GetProvider("test")
		require.NoError(t, err)
		assert.Equal(t, provider, registeredProvider)
	})

	t.Run("fails with empty name", func(t *testing.T) {
		provider := &MockProvider{}

		err := manager.RegisterProvider("", provider)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider name cannot be empty")
	})

	t.Run("fails with nil provider", func(t *testing.T) {
		err := manager.RegisterProvider("test", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider cannot be nil")
	})

	t.Run("fails when provider already registered", func(t *testing.T) {
		provider1 := &MockProvider{}
		provider2 := &MockProvider{}

		err := manager.RegisterProvider("duplicate", provider1)
		require.NoError(t, err)

		err = manager.RegisterProvider("duplicate", provider2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider \"duplicate\" already registered")
	})
}

func TestManagerImpl_GetProvider(t *testing.T) {
	manager := &ManagerImpl{
		providers: make(map[string]Provider),
		logger:    zap.NewNop(),
	}

	t.Run("returns registered provider", func(t *testing.T) {
		provider := &MockProvider{}
		manager.providers["test"] = provider

		result, err := manager.GetProvider("test")

		require.NoError(t, err)
		assert.Equal(t, provider, result)
	})

	t.Run("fails for non-existent provider", func(t *testing.T) {
		_, err := manager.GetProvider("nonexistent")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider \"nonexistent\" not found")
	})
}

func TestManagerImpl_GetSecretValue(t *testing.T) {
	ctx := context.Background()

	t.Run("retrieves secret value successfully", func(t *testing.T) {
		mockProvider := &MockProvider{}
		manager := &ManagerImpl{
			providers: map[string]Provider{
				"test": mockProvider,
			},
			logger: zap.NewNop(),
		}

		ref := Reference{
			Provider: "test",
			Path:     "app/secret",
			Key:      "password",
		}

		secret := &Secret{
			Path: "app/secret",
			Data: map[string][]byte{
				"password": []byte("secret123"),
				"username": []byte("admin"),
			},
		}

		mockProvider.On("GetSecret", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), "app/secret").Return(secret, nil)

		value, err := manager.GetSecretValue(ctx, ref)

		require.NoError(t, err)
		assert.Equal(t, []byte("secret123"), value)

		mockProvider.AssertExpectations(t)
	})

	t.Run("fails with invalid reference", func(t *testing.T) {
		mockProvider := &MockProvider{}
		manager := &ManagerImpl{
			providers: map[string]Provider{
				"test": mockProvider,
			},
			logger: zap.NewNop(),
		}

		ref := Reference{
			Provider: "", // Invalid - empty provider
			Path:     "app/secret",
			Key:      "password",
		}

		_, err := manager.GetSecretValue(ctx, ref)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid secret reference")
	})

	t.Run("fails when provider not found", func(t *testing.T) {
		mockProvider := &MockProvider{}
		manager := &ManagerImpl{
			providers: map[string]Provider{
				"test": mockProvider,
			},
			logger: zap.NewNop(),
		}

		ref := Reference{
			Provider: "nonexistent",
			Path:     "app/secret",
			Key:      "password",
		}

		_, err := manager.GetSecretValue(ctx, ref)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider \"nonexistent\" not found")
	})

	t.Run("fails when provider returns error", func(t *testing.T) {
		mockProvider := &MockProvider{}
		manager := &ManagerImpl{
			providers: map[string]Provider{
				"test": mockProvider,
			},
			logger: zap.NewNop(),
		}

		ref := Reference{
			Provider: "test",
			Path:     "app/secret",
			Key:      "password",
		}

		mockProvider.On("GetSecret", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), "app/secret").Return(nil, assert.AnError)

		_, err := manager.GetSecretValue(ctx, ref)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "getting secret from test")

		mockProvider.AssertExpectations(t)
	})

	t.Run("fails when key not found in secret", func(t *testing.T) {
		mockProvider := &MockProvider{}
		manager := &ManagerImpl{
			providers: map[string]Provider{
				"test": mockProvider,
			},
			logger: zap.NewNop(),
		}

		ref := Reference{
			Provider: "test",
			Path:     "app/secret",
			Key:      "nonexistent",
		}

		secret := &Secret{
			Path: "app/secret",
			Data: map[string][]byte{
				"password": []byte("secret123"),
			},
		}

		mockProvider.On("GetSecret", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), "app/secret").Return(secret, nil)

		_, err := manager.GetSecretValue(ctx, ref)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "key \"nonexistent\" not found in secret")

		mockProvider.AssertExpectations(t)
	})
}

func TestManagerImpl_GetSecret(t *testing.T) {
	mockProvider := &MockProvider{}
	manager := &ManagerImpl{
		providers: map[string]Provider{
			"test": mockProvider,
		},
		logger: zap.NewNop(),
	}

	ctx := context.Background()

	t.Run("retrieves secret successfully", func(t *testing.T) {
		secret := &Secret{
			Path: "app/secret",
			Data: map[string][]byte{
				"password": []byte("secret123"),
			},
		}

		mockProvider.On("GetSecret", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), "app/secret").Return(secret, nil)

		result, err := manager.GetSecret(ctx, "test", "app/secret")

		require.NoError(t, err)
		assert.Equal(t, secret, result)

		mockProvider.AssertExpectations(t)
	})

	t.Run("fails when provider not found", func(t *testing.T) {
		_, err := manager.GetSecret(ctx, "nonexistent", "app/secret")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider \"nonexistent\" not found")
	})
}

func TestManagerImpl_ListSecrets(t *testing.T) {
	mockProvider := &MockProvider{}
	manager := &ManagerImpl{
		providers: map[string]Provider{
			"test": mockProvider,
		},
		logger: zap.NewNop(),
	}

	ctx := context.Background()

	t.Run("lists secrets successfully", func(t *testing.T) {
		expectedPaths := []string{"app/secret1", "app/secret2"}

		mockProvider.On("ListSecrets", mock.MatchedBy(func(ctx context.Context) bool {
			return true
		}), "app").Return(expectedPaths, nil)

		paths, err := manager.ListSecrets(ctx, "test", "app")

		require.NoError(t, err)
		assert.Equal(t, expectedPaths, paths)

		mockProvider.AssertExpectations(t)
	})

	t.Run("fails when provider not found", func(t *testing.T) {
		_, err := manager.ListSecrets(ctx, "nonexistent", "app")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider \"nonexistent\" not found")
	})
}

func TestManagerImpl_ListProviders(t *testing.T) {
	manager := &ManagerImpl{
		providers: map[string]Provider{
			"file":  &MockProvider{},
			"vault": &MockProvider{},
		},
		logger: zap.NewNop(),
	}

	t.Run("lists all registered providers", func(t *testing.T) {
		providers := manager.ListProviders()

		assert.Len(t, providers, 2)
		assert.Contains(t, providers, "file")
		assert.Contains(t, providers, "vault")
	})

	t.Run("returns empty list when no providers", func(t *testing.T) {
		emptyManager := &ManagerImpl{
			providers: make(map[string]Provider),
			logger:    zap.NewNop(),
		}

		providers := emptyManager.ListProviders()

		assert.Empty(t, providers)
	})
}

// MockClosableProvider is a provider that implements Close()
type MockClosableProvider struct {
	MockProvider
}

func (m *MockClosableProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestManagerImpl_Close(t *testing.T) {
	t.Run("closes closable providers successfully", func(t *testing.T) {
		closableProvider := &MockClosableProvider{}
		regularProvider := &MockProvider{}

		manager := &ManagerImpl{
			providers: map[string]Provider{
				"closable": closableProvider,
				"regular":  regularProvider,
			},
			logger: zap.NewNop(),
		}

		closableProvider.On("Close").Return(nil)

		err := manager.Close()

		require.NoError(t, err)
		closableProvider.AssertExpectations(t)
	})

	t.Run("handles close errors", func(t *testing.T) {
		closableProvider := &MockClosableProvider{}

		manager := &ManagerImpl{
			providers: map[string]Provider{
				"closable": closableProvider,
			},
			logger: zap.NewNop(),
		}

		closableProvider.On("Close").Return(assert.AnError)

		err := manager.Close()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "errors closing providers")
		closableProvider.AssertExpectations(t)
	})

	t.Run("succeeds when no providers", func(t *testing.T) {
		manager := &ManagerImpl{
			providers: make(map[string]Provider),
			logger:    zap.NewNop(),
		}

		err := manager.Close()

		require.NoError(t, err)
	})
}

func TestManagerImpl_ThreadSafety(t *testing.T) {
	t.Run("concurrent access to providers map", func(t *testing.T) {
		manager := &ManagerImpl{
			providers: make(map[string]Provider),
			logger:    zap.NewNop(),
		}

		// Test concurrent RegisterProvider and ListProviders
		// This tests the RWMutex functionality
		done := make(chan bool)

		// Start a goroutine that repeatedly lists providers
		go func() {
			for range 100 {
				manager.ListProviders()
			}
			done <- true
		}()

		// Register providers concurrently
		for i := range 10 {
			provider := &MockProvider{}
			err := manager.RegisterProvider("test"+string(rune(i)), provider)
			require.NoError(t, err)
		}

		<-done

		// Verify final state
		providers := manager.ListProviders()
		assert.Len(t, providers, 10)
	})
}
