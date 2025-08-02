package secrets

import (
	"context"
	"fmt"
	"maps"
	"sync"

	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
)

// Manager defines the interface for secret management operations.
type Manager interface {
	// RegisterProvider registers a secret provider with a given name.
	RegisterProvider(name string, provider Provider) error

	// GetProvider returns a registered provider by name.
	GetProvider(name string) (Provider, error)

	// GetSecretValue retrieves a specific secret value using a reference.
	GetSecretValue(ctx context.Context, ref Reference) ([]byte, error)

	// GetSecret retrieves a full secret from a provider.
	GetSecret(ctx context.Context, providerName, path string) (*Secret, error)

	// ListSecrets lists secrets from a provider.
	ListSecrets(ctx context.Context, providerName, pathPrefix string) ([]string, error)

	// ListProviders returns the names of all registered providers.
	ListProviders() []string

	// Close closes all registered providers that implement io.Closer.
	Close() error
}

// ProviderFactory is a function that creates a provider based on configuration.
type ProviderFactory func(cfg *config.Config, logger *zap.Logger) (Provider, error)

// ManagerImpl manages multiple secret providers and provides a unified interface.
type ManagerImpl struct {
	providers map[string]Provider
	factories map[string]ProviderFactory
	mu        sync.RWMutex
	logger    *zap.Logger
}

// Ensure Manager implements ManagerInterface at compile time.
var _ Manager = (*ManagerImpl)(nil)

var (
	// Global registry of provider factories
	providerFactories = make(map[string]ProviderFactory)
	factoryMu         sync.RWMutex
)

// RegisterProviderFactory registers a provider factory globally.
func RegisterProviderFactory(name string, factory ProviderFactory) {
	factoryMu.Lock()
	defer factoryMu.Unlock()
	providerFactories[name] = factory
}

// GetProviderFactory returns a provider factory by name.
// This function is primarily intended for testing purposes.
func GetProviderFactory(name string) (ProviderFactory, bool) {
	factoryMu.RLock()
	defer factoryMu.RUnlock()
	factory, exists := providerFactories[name]
	return factory, exists
}

// NewManager creates a new secret manager and initializes providers based on configuration.
func NewManager(logger *zap.Logger, cfg *config.Config) (*ManagerImpl, error) {
	manager := &ManagerImpl{
		providers: make(map[string]Provider),
		factories: make(map[string]ProviderFactory),
		logger:    logger.With(zap.String("component", "secrets")),
	}

	// Copy factories from global registry
	factoryMu.RLock()
	maps.Copy(manager.factories, providerFactories)
	factoryMu.RUnlock()

	// Initialize file provider if enabled (OSS)
	if cfg.Secrets.Providers.File != nil && cfg.Secrets.Providers.File.Enabled {
		if factory, exists := manager.factories["file"]; exists {
			provider, err := factory(cfg, logger)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize file secret provider: %w", err)
			}
			if err := manager.RegisterProvider("file", provider); err != nil {
				return nil, fmt.Errorf("failed to register file secret provider: %w", err)
			}

			logger.Info("registered file secret provider",
				zap.String("base_path", cfg.Secrets.Providers.File.BasePath))
		} else {
			return nil, fmt.Errorf("file provider factory not registered")
		}
	}

	// Initialize Vault provider if enabled (Pro feature)
	if cfg.Secrets.Providers.Vault != nil && cfg.Secrets.Providers.Vault.Enabled {
		if factory, exists := manager.factories["vault"]; exists {
			provider, err := factory(cfg, logger)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize vault secret provider: %w", err)
			}

			if err := manager.RegisterProvider("vault", provider); err != nil {
				return nil, fmt.Errorf("failed to register vault secret provider: %w", err)
			}

			logger.Info("registered vault secret provider",
				zap.String("address", cfg.Secrets.Providers.Vault.Address))
		} else {
			return nil, fmt.Errorf("vault provider factory not registered")
		}
	}

	return manager, nil
}

// RegisterProvider registers a secret provider with a given name.
func (m *ManagerImpl) RegisterProvider(name string, provider Provider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	if _, exists := m.providers[name]; exists {
		return fmt.Errorf("provider %q already registered", name)
	}

	m.providers[name] = provider
	m.logger.Info("registered secret provider", zap.String("provider", name))

	return nil
}

// GetProvider returns a registered provider by name.
func (m *ManagerImpl) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %q not found", name)
	}

	return provider, nil
}

// GetSecretValue retrieves a specific secret value using a reference.
func (m *ManagerImpl) GetSecretValue(ctx context.Context, ref Reference) ([]byte, error) {
	if err := ref.Validate(); err != nil {
		return nil, fmt.Errorf("invalid secret reference: %w", err)
	}

	provider, err := m.GetProvider(ref.Provider)
	if err != nil {
		return nil, err
	}

	m.logger.Debug("retrieving secret",
		zap.String("provider", ref.Provider),
		zap.String("path", ref.Path),
		zap.String("key", ref.Key))

	secret, err := provider.GetSecret(ctx, ref.Path)
	if err != nil {
		return nil, fmt.Errorf("getting secret from %s: %w", ref.Provider, err)
	}

	value, ok := secret.GetValue(ref.Key)
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret at path %q", ref.Key, ref.Path)
	}

	return value, nil
}

// GetSecret retrieves a full secret from a provider.
func (m *ManagerImpl) GetSecret(ctx context.Context, providerName, path string) (*Secret, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.GetSecret(ctx, path)
}

// ListSecrets lists secrets from a provider.
func (m *ManagerImpl) ListSecrets(ctx context.Context, providerName, pathPrefix string) ([]string, error) {
	provider, err := m.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	return provider.ListSecrets(ctx, pathPrefix)
}

// ListProviders returns the names of all registered providers.
func (m *ManagerImpl) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.providers))
	for name := range m.providers {
		names = append(names, name)
	}
	return names
}

// Close closes all registered providers that implement io.Closer.
func (m *ManagerImpl) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, provider := range m.providers {
		if closer, ok := provider.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, fmt.Errorf("closing provider %s: %w", name, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}

	return nil
}
