package environments

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/coss/license"
	"go.flipt.io/flipt/internal/credentials"
	"go.flipt.io/flipt/internal/product"
	"go.flipt.io/flipt/internal/secrets"
	"go.flipt.io/flipt/internal/storage"
	envsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	storagegit "go.flipt.io/flipt/internal/storage/git"
	"go.uber.org/zap/zaptest"
)

func Test_NewRepositoryManager(t *testing.T) {
	var (
		logger         = zaptest.NewLogger(t)
		cfg            = &config.Config{}
		secretsManager = &secrets.MockManager{}
		licenseManager = &license.MockManager{}

		rm = NewRepositoryManager(logger, cfg, secretsManager, licenseManager)
	)

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
		expectCached  bool
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
			setupMocks:   func(lm *license.MockManager, sm *secrets.MockManager) {},
			expectCached: true,
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

			ctx := t.Context()
			repo, err := rm.GetOrCreate(ctx, tt.envConf, tt.storageConfig, credentials)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, repo)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, repo)

				// Verify repository is cached
				if tt.expectCached {
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

			ctx := t.Context()
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

			ctx := t.Context()
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
	ctx := t.Context()
	err := subscriber.Notify(ctx, map[string]string{"main": "abc123"})
	require.NoError(t, err)
}

func Test_NewStore_DiscoveredBranchDoesNotReplaceEnvironment(t *testing.T) {
	var (
		ctx    = t.Context()
		logger = zaptest.NewLogger(t)
		cfg    = &config.Config{
			Environments: map[string]*config.EnvironmentConfig{
				"production": {
					Name:      "production",
					Storage:   "default",
					Directory: "production",
					Default:   true,
				},
				"staging": {
					Name:      "staging",
					Storage:   "default",
					Directory: "staging",
				},
			},
			Storage: map[string]*config.StorageConfig{
				"default": {
					Backend: config.StorageBackendConfig{
						Type: config.MemoryStorageBackendType,
					},
					Branch: "main",
				},
			},
		}
	)

	store, err := NewStore(ctx, logger, cfg, &secrets.MockManager{}, &license.MockManager{})
	require.NoError(t, err)

	production, err := store.Get(ctx, "production")
	require.NoError(t, err)

	staging, err := store.Get(ctx, "staging")
	require.NoError(t, err)

	repo := staging.(interface{ Repository() *storagegit.Repository }).Repository()

	// a branch pushed directly to the repository under the staging branch prefix
	// whose name collides with the static production environment
	const rogue = "flipt/staging/production"
	require.NoError(t, repo.CreateBranchIfNotExists(ctx, rogue, storagegit.WithBase("main")))

	// committing to the branch notifies subscribers, which discovers it
	_, err = repo.UpdateAndPush(ctx, rogue, func(fs envsfs.Filesystem) (string, error) {
		fi, err := fs.OpenFile("staging/README.md", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
		if err != nil {
			return "", err
		}
		return "rogue commit", fi.Close()
	})
	require.NoError(t, err)

	got, err := store.Get(ctx, "production")
	require.NoError(t, err)
	assert.Same(t, production, got, "discovered branch replaced the production environment")
	assert.Nil(t, got.Configuration().Base)

	// removing the branch and pruning it must not remove the production environment
	require.NoError(t, repo.DeleteBranch(ctx, rogue))
	require.NoError(t, repo.Storer.RemoveReference(plumbing.NewRemoteReferenceName("origin", rogue)))

	_, err = repo.UpdateAndPush(ctx, "main", func(fs envsfs.Filesystem) (string, error) {
		fi, err := fs.OpenFile("staging/README.md", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
		if err != nil {
			return "", err
		}
		return "trigger refresh", fi.Close()
	})
	require.NoError(t, err)

	got, err = store.Get(ctx, "production")
	require.NoError(t, err)
	assert.Same(t, production, got, "pruned branch removed the production environment")
}

const invalidFeatureYAML = "version: \"1.6\"\nflags: \"not-a-list\"\n"

func booleanFlagYAML(key, name string) string {
	return fmt.Sprintf(`version: "1.6"
namespace:
  key: default
  name: Default
flags:
  - key: %s
    name: %s
    type: BOOLEAN_FLAG_TYPE
    enabled: true
    variants: []
    rules: []
    rollouts: []
segments: []
`, key, name)
}

func pushFile(t *testing.T, repo *storagegit.Repository, branch, path, contents, message string) plumbing.Hash {
	t.Helper()

	hash, err := repo.UpdateAndPush(t.Context(), branch, func(fs envsfs.Filesystem) (string, error) {
		fi, err := fs.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
		if err != nil {
			return "", err
		}
		if _, err := fi.Write([]byte(contents)); err != nil {
			_ = fi.Close()
			return "", err
		}
		return message, fi.Close()
	})
	require.NoError(t, err)

	return hash
}

type flagReader interface {
	EvaluationStore() (storage.ReadOnlyStore, error)
}

func requireEnvironmentFlag(t *testing.T, env flagReader, key, name string) {
	t.Helper()

	eval, err := env.EvaluationStore()
	require.NoError(t, err)

	flag, err := eval.GetFlag(t.Context(), storage.NewResource("default", key))
	require.NoError(t, err)
	assert.Equal(t, name, flag.Name)
}

func Test_NewStore_MalformedBranchDoesNotBlockRecovery(t *testing.T) {
	var (
		ctx    = t.Context()
		logger = zaptest.NewLogger(t)
		cfg    = &config.Config{
			Environments: map[string]*config.EnvironmentConfig{
				"production": {
					Name:    "production",
					Storage: "default",
					Default: true,
				},
			},
			Storage: map[string]*config.StorageConfig{
				"default": {
					Backend: config.StorageBackendConfig{
						Type: config.MemoryStorageBackendType,
					},
					Branch: "main",
				},
			},
		}
	)

	store, err := NewStore(ctx, logger, cfg, &secrets.MockManager{}, &license.MockManager{})
	require.NoError(t, err)

	production, err := store.Get(ctx, "production")
	require.NoError(t, err)

	repo := production.(interface{ Repository() *storagegit.Repository }).Repository()

	const (
		broken    = "flipt/production/broken"
		healthy   = "flipt/production/healthy"
		temporary = "flipt/production/temporary"
		features  = "default/features.yaml"
	)

	require.NoError(t, repo.CreateBranchIfNotExists(ctx, broken, storagegit.WithBase("main")))
	pushFile(t, repo, broken, features, invalidFeatureYAML, "invalid features")

	_, err = store.Get(ctx, "broken")
	require.Error(t, err)
	var notFound errors.ErrNotFound
	require.ErrorAs(t, err, &notFound)

	require.NoError(t, repo.CreateBranchIfNotExists(ctx, healthy, storagegit.WithBase("main")))
	healthyHash := pushFile(t, repo, healthy, features, booleanFlagYAML("healthy_flag", "Healthy Flag"), "healthy features")

	healthyEnv, err := store.Get(ctx, "healthy")
	require.NoError(t, err)
	requireEnvironmentFlag(t, healthyEnv, "healthy_flag", "Healthy Flag")
	ns, err := healthyEnv.GetNamespace(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, healthyHash.String(), ns.Revision)
	require.NotNil(t, healthyEnv.Configuration().Base)
	assert.Equal(t, "production", *healthyEnv.Configuration().Base)

	_, err = store.Get(ctx, "broken")
	require.ErrorAs(t, err, &notFound)

	require.NoError(t, repo.CreateBranchIfNotExists(ctx, temporary, storagegit.WithBase("main")))
	pushFile(t, repo, temporary, features, booleanFlagYAML("temp_flag", "Temp Flag"), "temporary features")
	_, err = store.Get(ctx, "temporary")
	require.NoError(t, err)

	// Deleting temporary while broken is still invalid must still prune it.
	require.NoError(t, repo.DeleteBranch(ctx, temporary))
	require.NoError(t, repo.Storer.RemoveReference(plumbing.NewRemoteReferenceName("origin", temporary)))
	pushFile(t, repo, "main", "refresh.txt", "refresh", "trigger refresh")

	_, err = store.Get(ctx, "temporary")
	require.ErrorAs(t, err, &notFound)

	healthyEnv, err = store.Get(ctx, "healthy")
	require.NoError(t, err)
	requireEnvironmentFlag(t, healthyEnv, "healthy_flag", "Healthy Flag")

	_, err = store.Get(ctx, "broken")
	require.ErrorAs(t, err, &notFound)

	got, err := store.Get(ctx, "production")
	require.NoError(t, err)
	assert.Same(t, production, got)

	brokenHash := pushFile(t, repo, broken, features, booleanFlagYAML("broken_flag", "Broken Flag"), "repair broken")

	brokenEnv, err := store.Get(ctx, "broken")
	require.NoError(t, err)
	requireEnvironmentFlag(t, brokenEnv, "broken_flag", "Broken Flag")
	ns, err = brokenEnv.GetNamespace(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, brokenHash.String(), ns.Revision)
	require.NotNil(t, brokenEnv.Configuration().Base)
	assert.Equal(t, "production", *brokenEnv.Configuration().Base)

	healthyEnv, err = store.Get(ctx, "healthy")
	require.NoError(t, err)
	requireEnvironmentFlag(t, healthyEnv, "healthy_flag", "Healthy Flag")
	ns, err = healthyEnv.GetNamespace(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, healthyHash.String(), ns.Revision)

	healthyEval, err := healthyEnv.EvaluationStore()
	require.NoError(t, err)
	_, err = healthyEval.GetFlag(ctx, storage.NewResource("default", "broken_flag"))
	require.Error(t, err)

	brokenEval, err := brokenEnv.EvaluationStore()
	require.NoError(t, err)
	_, err = brokenEval.GetFlag(ctx, storage.NewResource("default", "healthy_flag"))
	require.Error(t, err)

	pushFile(t, repo, "main", "refresh.txt", "refresh again", "trigger unchanged refresh")

	again, err := store.Get(ctx, "broken")
	require.NoError(t, err)
	assert.Same(t, brokenEnv, again)
	requireEnvironmentFlag(t, again, "broken_flag", "Broken Flag")
	ns, err = again.GetNamespace(ctx, "default")
	require.NoError(t, err)
	assert.Equal(t, brokenHash.String(), ns.Revision)
}
