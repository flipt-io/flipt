package signing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.flipt.io/build/internal/dagger"
)

// VaultConfig represents the configuration for a test Vault instance
type VaultConfig struct {
	RootToken   string
	GPGKeyPath  string
	GPGKeyData  map[string]string
	VaultAddr   string
	VaultPort   int
}

// DefaultVaultConfig returns a default configuration for test Vault
func DefaultVaultConfig() *VaultConfig {
	return &VaultConfig{
		RootToken:  "test-root-token",
		GPGKeyPath: "secret/flipt/signing-key",
		VaultAddr:  "http://vault:8200",
		VaultPort:  8200,
	}
}

// WithGPGKey sets up the vault test case with a pre-configured GPG key
func WithGPGKey(privateKey, publicKey string) func(*VaultConfig) {
	return func(cfg *VaultConfig) {
		cfg.GPGKeyData = map[string]string{
			"private_key": privateKey,
			"public_key":  publicKey,
		}
	}
}

// SetupVaultContainer creates and configures a Vault container for testing
func SetupVaultContainer(ctx context.Context, client *dagger.Client, cfg *VaultConfig) (*dagger.Service, error) {
	// Create Vault container
	vault := client.Container().
		From("hashicorp/vault:1.17.5").
		WithEnvVariable("VAULT_DEV_ROOT_TOKEN_ID", cfg.RootToken).
		WithEnvVariable("VAULT_DEV_LISTEN_ADDRESS", "0.0.0.0:8200").
		WithEnvVariable("VAULT_LOG_LEVEL", "debug").
		WithExposedPort(cfg.VaultPort).
		WithExec([]string{"vault", "server", "-dev"}).
		AsService()

	// Wait for Vault to be ready and then configure it
	setupContainer := client.Container().
		From("hashicorp/vault:1.17.5").
		WithEnvVariable("VAULT_ADDR", cfg.VaultAddr).
		WithEnvVariable("VAULT_TOKEN", cfg.RootToken).
		WithServiceBinding("vault", vault)

	// Enable the KV v2 secrets engine
	_, err := setupContainer.
		WithExec([]string{"vault", "secrets", "enable", "-version=2", "kv"}).
		Sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to enable KV v2 engine: %w", err)
	}

	// Store the GPG key if provided
	if len(cfg.GPGKeyData) > 0 {
		secretData, err := json.Marshal(cfg.GPGKeyData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal secret data: %w", err)
		}

		// Create a temporary file with the secret data
		secretFile := fmt.Sprintf("/tmp/secret-%d.json", time.Now().UnixNano())
		
		_, err = setupContainer.
			WithNewFile(secretFile, string(secretData)).
			WithExec([]string{"vault", "kv", "put", cfg.GPGKeyPath, fmt.Sprintf("@%s", secretFile)}).
			Sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to store GPG key in Vault: %w", err)
		}
	}

	return vault, nil
}

// VaultTestCase represents a test case that uses Vault
type VaultTestCase struct {
	Name        string
	VaultConfig *VaultConfig
	FliptEnvs   map[string]string
}

// NewVaultTestCase creates a new Vault-based test case
func NewVaultTestCase(name string, opts ...func(*VaultConfig)) *VaultTestCase {
	cfg := DefaultVaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return &VaultTestCase{
		Name:        name,
		VaultConfig: cfg,
		FliptEnvs:   make(map[string]string),
	}
}

// WithFliptEnv adds environment variables for Flipt configuration
func (tc *VaultTestCase) WithFliptEnv(key, value string) *VaultTestCase {
	tc.FliptEnvs[key] = value
	return tc
}

// WithVaultSecrets configures Flipt to use Vault for secrets
func (tc *VaultTestCase) WithVaultSecrets() *VaultTestCase {
	return tc.
		WithFliptEnv("FLIPT_SECRETS_PROVIDERS_VAULT_ENABLED", "true").
		WithFliptEnv("FLIPT_SECRETS_PROVIDERS_VAULT_ADDRESS", tc.VaultConfig.VaultAddr).
		WithFliptEnv("FLIPT_SECRETS_PROVIDERS_VAULT_AUTH_METHOD", "token").
		WithFliptEnv("FLIPT_SECRETS_PROVIDERS_VAULT_TOKEN", tc.VaultConfig.RootToken).
		WithFliptEnv("FLIPT_SECRETS_PROVIDERS_VAULT_MOUNT", "secret")
}

// WithCommitSigning configures Flipt to use commit signing
func (tc *VaultTestCase) WithCommitSigning() *VaultTestCase {
	return tc.
		WithFliptEnv("FLIPT_STORAGE_DEFAULT_SIGNATURE_ENABLED", "true").
		WithFliptEnv("FLIPT_STORAGE_DEFAULT_SIGNATURE_TYPE", "gpg").
		WithFliptEnv("FLIPT_STORAGE_DEFAULT_SIGNATURE_KEY_REF_PROVIDER", "vault").
		WithFliptEnv("FLIPT_STORAGE_DEFAULT_SIGNATURE_KEY_REF_PATH", tc.VaultConfig.GPGKeyPath).
		WithFliptEnv("FLIPT_STORAGE_DEFAULT_SIGNATURE_KEY_REF_KEY", "private_key").
		WithFliptEnv("FLIPT_STORAGE_DEFAULT_SIGNATURE_NAME", "Flipt Test Bot").
		WithFliptEnv("FLIPT_STORAGE_DEFAULT_SIGNATURE_EMAIL", "test-bot@flipt.io").
		WithFliptEnv("FLIPT_STORAGE_DEFAULT_SIGNATURE_KEY_ID", "test-bot@flipt.io")
}

// ApplyToContainer applies the test case configuration to a Flipt container
func (tc *VaultTestCase) ApplyToContainer(flipt *dagger.Container, vault *dagger.Service) *dagger.Container {
	// Bind Vault service
	flipt = flipt.WithServiceBinding("vault", vault)

	// Apply all environment variables
	for key, value := range tc.FliptEnvs {
		flipt = flipt.WithEnvVariable(key, value)
	}

	return flipt
}