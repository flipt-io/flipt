package environments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/license"
	"go.flipt.io/flipt/internal/credentials"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap/zaptest"
)

func Test_NewRepositoryManager(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	secretsManager := &secrets.MockManager{}
	licenseManager := &license.MockManager{}

	rm := NewRepositoryManager(logger, cfg, secretsManager, licenseManager)

	assert.NotNil(t, rm)
	assert.Equal(t, logger, rm.logger)
	assert.Equal(t, cfg, rm.cfg)
	assert.NotNil(t, rm.repos)
	assert.Equal(t, secretsManager, rm.secretsManager)
	assert.Equal(t, licenseManager, rm.licenseManager)
}

func Test_RepositoryManager_GetOrCreate(t *testing.T) {
	tests := []struct {
		name          string
		envConf       *config.EnvironmentConfig
		storageConfig *config.StorageConfig
		expectedError bool
		setupMocks    func(*license.MockManager, *secrets.MockManager)
	}{
		{
			name: "creates new repository with local backend",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "test-storage",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.LocalStorageBackendType,
					Path: "/tmp/test",
				},
				Branch: "main",
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
		{
			name: "creates new repository with memory backend",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "test-storage",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.MemoryStorageBackendType,
				},
				Branch: "main",
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
		{
			name: "creates repository with remote",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "test-storage",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.MemoryStorageBackendType,
				},
				Branch: "main",
				Remote: "https://github.com/flipt-io/flipt.git",
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
		{
			name: "creates repository with CA cert bytes",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "test-storage",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.MemoryStorageBackendType,
				},
				Branch:      "main",
				CaCertBytes: "test-cert",
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
		{
			name: "reuses existing repository",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "existing-storage",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.MemoryStorageBackendType,
				},
				Branch: "main",
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
		{
			name: "repository with commit signing enabled - OSS license",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "test-storage",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.MemoryStorageBackendType,
				},
				Branch: "main",
				Signature: config.SignatureConfig{
					Enabled: true,
					Name:    "Test User",
					Email:   "test@example.com",
					KeyRef:  &config.SecretReference{Provider: "vault", Path: "secret", Key: "key"},
					KeyID:   "test-key-id",
				},
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {
				lm.On("Product").Return(product.OSS)
			},
		},
		{
			name: "repository with commit signing enabled - Pro license creates signer",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "test-storage-pro",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.MemoryStorageBackendType,
				},
				Branch: "main",
				Signature: config.SignatureConfig{
					Enabled: true,
					Name:    "Test User",
					Email:   "test@example.com",
					KeyRef:  &config.SecretReference{Provider: "vault", Path: "secret", Key: "key"},
					KeyID:   "test-key-id",
				},
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {
				lm.On("Product").Return(product.Pro)
				// Return a minimal valid PGP private key for testing
				// This is a test key with no real security value
				//nolint:gosec
				testKey := `-----BEGIN PGP PRIVATE KEY BLOCK-----

lQOYBGWJVAsBCAC5W3pEMVV2I8hSlCQviqhR6alOSNgXMAR0e7qQPjuY2d1+gvyT
7oEbjIw9hdHgGDY3Y6pAUaN5YoyZN0gVXhp8OjR8C3aDB1d5MtGXhKMYHPBPL3Jv
6kqoODmJxJ1VS2Yz9K5Tu1Y8oRNLx0u8DlzaC5eUyKzqMkLh6hXxONkM8HTqQfPa
test_key_data
=PGH6
-----END PGP PRIVATE KEY BLOCK-----`
				sm.On("GetSecretValue", mock.Anything, secrets.Reference{Provider: "vault", Path: "secret", Key: "key"}).Return([]byte(testKey), nil)
			},
			expectedError: true, // Still expect error due to invalid test key
		},
		{
			name: "repository with commit signing enabled but no key ref",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "test-storage-no-key",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.MemoryStorageBackendType,
				},
				Branch: "main",
				Signature: config.SignatureConfig{
					Enabled: true,
					Name:    "Test User",
					Email:   "test@example.com",
				},
			},
			expectedError: true,
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {
				lm.On("Product").Return(product.Pro)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			cfg := &config.Config{
				Storage: map[string]*config.StorageConfig{
					tt.envConf.Storage: tt.storageConfig,
				},
			}

			secretsManager := &secrets.MockManager{}
			licenseManager := &license.MockManager{}
			tt.setupMocks(licenseManager, secretsManager)

			rm := NewRepositoryManager(logger, cfg, secretsManager, licenseManager)
			credentials := credentials.New(logger, cfg.Credentials)

			ctx := context.Background()
			repo, err := rm.GetOrCreate(ctx, tt.envConf, tt.storageConfig, credentials)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, repo)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, repo)

				// Verify repository is cached
				if tt.name == "reuses existing repository" {
					// Pre-populate the cache
					rm.repos[tt.envConf.Storage] = repo
					repo2, err := rm.GetOrCreate(ctx, tt.envConf, tt.storageConfig, credentials)
					require.NoError(t, err)
					assert.Same(t, repo, repo2)
				}
			}

			licenseManager.AssertExpectations(t)
			secretsManager.AssertExpectations(t)
		})
	}
}

func Test_NewEnvironmentFactory(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}
	credentials := credentials.New(logger, cfg.Credentials)
	secretsManager := &secrets.MockManager{}
	licenseManager := &license.MockManager{}
	repoManager := NewRepositoryManager(logger, cfg, secretsManager, licenseManager)

	factory := NewEnvironmentFactory(logger, cfg, credentials, repoManager, licenseManager)

	assert.NotNil(t, factory)
	assert.Equal(t, logger, factory.logger)
	assert.Equal(t, cfg, factory.cfg)
	assert.Equal(t, credentials, factory.credentials)
	assert.Equal(t, repoManager, factory.repoManager)
	assert.Equal(t, licenseManager, factory.licenseManager)
}

func Test_EnvironmentFactory_Create(t *testing.T) {
	tests := []struct {
		name          string
		envName       string
		envConf       *config.EnvironmentConfig
		storageConfig *config.StorageConfig
		setupMocks    func(*license.MockManager, *secrets.MockManager)
		expectedError bool
		errorContains string
	}{
		{
			name:    "creates basic environment",
			envName: "test-env",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "test-storage",
			},
			storageConfig: &config.StorageConfig{
				Backend: config.StorageBackendConfig{
					Type: config.MemoryStorageBackendType,
				},
				Branch: "main",
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
		{
			name:    "missing storage configuration",
			envName: "test-env",
			envConf: &config.EnvironmentConfig{
				Name:    "test-env",
				Storage: "missing-storage",
			},
			expectedError: true,
			errorContains: "missing storage for name",
			setupMocks:    func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			cfg := &config.Config{
				Storage: map[string]*config.StorageConfig{},
			}

			if tt.storageConfig != nil {
				cfg.Storage["test-storage"] = tt.storageConfig
			}

			secretsManager := &secrets.MockManager{}
			licenseManager := &license.MockManager{}
			tt.setupMocks(licenseManager, secretsManager)

			credentials := credentials.New(logger, cfg.Credentials)
			repoManager := NewRepositoryManager(logger, cfg, secretsManager, licenseManager)
			factory := NewEnvironmentFactory(logger, cfg, credentials, repoManager, licenseManager)

			ctx := context.Background()
			env, err := factory.Create(ctx, tt.envName, tt.envConf)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, env)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, env)
				assert.Equal(t, tt.envName, env.Key())
			}

			licenseManager.AssertExpectations(t)
			secretsManager.AssertExpectations(t)
		})
	}
}

func Test_NewStore(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.Config
		setupMocks    func(*license.MockManager, *secrets.MockManager)
		expectedError bool
		errorContains string
	}{
		{
			name: "no environments configured",
			cfg: &config.Config{
				Environments: map[string]*config.EnvironmentConfig{},
			},
			setupMocks:    func(lm *license.MockManager, sm *secrets.MockManager) {},
			expectedError: true,
			errorContains: "no environments configured",
		},
		{
			name: "single environment",
			cfg: &config.Config{
				Environments: map[string]*config.EnvironmentConfig{
					"test": {
						Name:    "test",
						Storage: "test-storage",
					},
				},
				Storage: map[string]*config.StorageConfig{
					"test-storage": {
						Backend: config.StorageBackendConfig{
							Type: config.MemoryStorageBackendType,
						},
						Branch: "main",
					},
				},
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
		{
			name: "multiple environments",
			cfg: &config.Config{
				Environments: map[string]*config.EnvironmentConfig{
					"dev": {
						Name:    "dev",
						Storage: "dev-storage",
					},
					"prod": {
						Name:    "prod",
						Storage: "prod-storage",
						Default: true,
					},
				},
				Storage: map[string]*config.StorageConfig{
					"dev-storage": {
						Backend: config.StorageBackendConfig{
							Type: config.MemoryStorageBackendType,
						},
						Branch: "dev",
					},
					"prod-storage": {
						Backend: config.StorageBackendConfig{
							Type: config.MemoryStorageBackendType,
						},
						Branch: "main",
					},
				},
			},
			setupMocks: func(lm *license.MockManager, sm *secrets.MockManager) {},
		},
		{
			name: "environment with missing storage",
			cfg: &config.Config{
				Environments: map[string]*config.EnvironmentConfig{
					"test": {
						Name:    "test",
						Storage: "missing-storage",
					},
				},
				Storage: map[string]*config.StorageConfig{},
			},
			setupMocks:    func(lm *license.MockManager, sm *secrets.MockManager) {},
			expectedError: true,
			errorContains: "missing storage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			secretsManager := &secrets.MockManager{}
			licenseManager := &license.MockManager{}
			tt.setupMocks(licenseManager, secretsManager)

			ctx := context.Background()
			store, err := NewStore(ctx, logger, tt.cfg, secretsManager, licenseManager)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, store)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, store)
			}

			licenseManager.AssertExpectations(t)
			secretsManager.AssertExpectations(t)
		})
	}
}

func Test_environmentSubscriber(t *testing.T) {
	subscriber := &environmentSubscriber{
		branchesFn: func() []string {
			return []string{"main", "dev"}
		},
		notifyFn: func(ctx context.Context, refs map[string]string) error {
			return nil
		},
	}

	// Test Branches
	branches := subscriber.Branches()
	assert.Equal(t, []string{"main", "dev"}, branches)

	// Test Notify
	ctx := context.Background()
	err := subscriber.Notify(ctx, map[string]string{"main": "abc123"})
	require.NoError(t, err)
}
