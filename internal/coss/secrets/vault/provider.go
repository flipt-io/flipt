package vault

import (
	"context"
	"fmt"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

func init() {
	// Register vault provider factory
	secrets.RegisterProviderFactory("vault", func(cfg *config.Config, logger *zap.Logger) (secrets.Provider, error) {
		if cfg.Secrets.Providers.Vault == nil {
			return nil, fmt.Errorf("vault provider configuration not found")
		}
		
		vaultConfig := Config{
			Address:    cfg.Secrets.Providers.Vault.Address,
			AuthMethod: cfg.Secrets.Providers.Vault.AuthMethod,
			Role:       cfg.Secrets.Providers.Vault.Role,
			Mount:      cfg.Secrets.Providers.Vault.Mount,
			Token:      cfg.Secrets.Providers.Vault.Token,
			Namespace:  cfg.Secrets.Providers.Vault.Namespace,
		}
		
		return NewProvider(vaultConfig, logger)
	})
}

// Provider implements secrets.Provider for HashiCorp Vault.
type Provider struct {
	client *vault.Client
	mount  string
	logger *zap.Logger
}

// Config contains configuration for the Vault provider.
type Config struct {
	Address    string
	AuthMethod string
	Role       string
	Mount      string
	Token      string
	Namespace  string
}

// NewProvider creates a new Vault secret provider.
func NewProvider(cfg Config, logger *zap.Logger) (*Provider, error) {
	config := vault.DefaultConfig()
	config.Address = cfg.Address
	
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	// Set namespace if provided
	if cfg.Namespace != "" {
		client.SetNamespace(cfg.Namespace)
	}

	// Configure authentication
	if err := authenticate(client, cfg); err != nil {
		return nil, fmt.Errorf("authenticating with vault: %w", err)
	}

	return &Provider{
		client: client,
		mount:  cfg.Mount,
		logger: logger,
	}, nil
}

// GetSecret retrieves a secret from Vault.
func (p *Provider) GetSecret(ctx context.Context, path string) (*secrets.Secret, error) {
	secretPath := fmt.Sprintf("%s/data/%s", p.mount, path)
	
	p.logger.Debug("reading secret from vault",
		zap.String("path", path),
		zap.String("mount", p.mount))

	secret, err := p.client.Logical().ReadWithContext(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("reading from vault: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}

	// Extract data from the versioned secret format
	data := make(map[string][]byte)
	metadata := make(map[string]string)
	version := ""

	// Vault KV v2 wraps the actual data
	if secretData, ok := secret.Data["data"].(map[string]interface{}); ok {
		for k, v := range secretData {
			switch val := v.(type) {
			case string:
				data[k] = []byte(val)
			case []byte:
				data[k] = val
			default:
				// Skip non-string/byte values
				p.logger.Warn("skipping non-string value in secret",
					zap.String("key", k),
					zap.String("type", fmt.Sprintf("%T", v)))
			}
		}
	}

	// Extract metadata
	if meta, ok := secret.Data["metadata"].(map[string]interface{}); ok {
		for k, v := range meta {
			if strVal, ok := v.(string); ok {
				metadata[k] = strVal
			}
		}
		if v, ok := meta["version"].(float64); ok {
			version = fmt.Sprintf("%.0f", v)
		}
	}

	return &secrets.Secret{
		Path:     path,
		Data:     data,
		Metadata: metadata,
		Version:  version,
	}, nil
}

// PutSecret stores a secret in Vault.
func (p *Provider) PutSecret(ctx context.Context, path string, secret *secrets.Secret) error {
	secretPath := fmt.Sprintf("%s/data/%s", p.mount, path)
	
	// Convert byte data to strings for Vault
	data := make(map[string]interface{})
	for k, v := range secret.Data {
		data[k] = string(v)
	}

	// Wrap data for KV v2
	payload := map[string]interface{}{
		"data": data,
	}

	// Add metadata if present
	if len(secret.Metadata) > 0 {
		options := make(map[string]interface{})
		for k, v := range secret.Metadata {
			options[k] = v
		}
		payload["options"] = options
	}

	p.logger.Debug("writing secret to vault",
		zap.String("path", path),
		zap.String("mount", p.mount))

	_, err := p.client.Logical().WriteWithContext(ctx, secretPath, payload)
	if err != nil {
		return fmt.Errorf("writing to vault: %w", err)
	}

	return nil
}

// DeleteSecret removes a secret from Vault.
func (p *Provider) DeleteSecret(ctx context.Context, path string) error {
	// For KV v2, we need to use the metadata path for deletion
	deletePath := fmt.Sprintf("%s/metadata/%s", p.mount, path)
	
	p.logger.Debug("deleting secret from vault",
		zap.String("path", path),
		zap.String("mount", p.mount))

	_, err := p.client.Logical().DeleteWithContext(ctx, deletePath)
	if err != nil {
		return fmt.Errorf("deleting from vault: %w", err)
	}

	return nil
}

// ListSecrets returns all secret paths matching the prefix.
func (p *Provider) ListSecrets(ctx context.Context, pathPrefix string) ([]string, error) {
	listPath := fmt.Sprintf("%s/metadata/%s", p.mount, pathPrefix)
	
	p.logger.Debug("listing secrets from vault",
		zap.String("prefix", pathPrefix),
		zap.String("mount", p.mount))

	secret, err := p.client.Logical().ListWithContext(ctx, listPath)
	if err != nil {
		// Vault returns 404 for empty directories
		if strings.Contains(err.Error(), "404") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("listing from vault: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	var paths []string
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

	return paths, nil
}

// authenticate handles Vault authentication based on the configured method.
func authenticate(client *vault.Client, cfg Config) error {
	switch cfg.AuthMethod {
	case "token":
		if cfg.Token == "" {
			// Try to read from environment or token file
			token := os.Getenv("VAULT_TOKEN")
			if token == "" {
				return fmt.Errorf("no vault token provided")
			}
			cfg.Token = token
		}
		client.SetToken(cfg.Token)
		return nil

	case "kubernetes":
		return authenticateKubernetes(client, cfg.Role)

	case "approle":
		return authenticateAppRole(client, cfg.Role)

	default:
		return fmt.Errorf("unsupported auth method: %s", cfg.AuthMethod)
	}
}

// authenticateKubernetes performs Kubernetes authentication.
func authenticateKubernetes(client *vault.Client, role string) error {
	// Read service account token
	tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return fmt.Errorf("reading kubernetes service account token: %w", err)
	}

	// Login with Kubernetes auth
	data := map[string]interface{}{
		"role": role,
		"jwt":  string(tokenBytes),
	}

	path := "auth/kubernetes/login"
	secret, err := client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("kubernetes auth login: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth info returned from kubernetes login")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}

// authenticateAppRole performs AppRole authentication.
func authenticateAppRole(client *vault.Client, role string) error {
	// Read role ID and secret ID from environment
	roleID := os.Getenv("VAULT_ROLE_ID")
	secretID := os.Getenv("VAULT_SECRET_ID")

	if roleID == "" || secretID == "" {
		return fmt.Errorf("VAULT_ROLE_ID and VAULT_SECRET_ID must be set for approle auth")
	}

	data := map[string]interface{}{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	path := "auth/approle/login"
	secret, err := client.Logical().Write(path, data)
	if err != nil {
		return fmt.Errorf("approle auth login: %w", err)
	}

	if secret == nil || secret.Auth == nil {
		return fmt.Errorf("no auth info returned from approle login")
	}

	client.SetToken(secret.Auth.ClientToken)
	return nil
}

// Close implements io.Closer for cleanup.
func (p *Provider) Close() error {
	// Vault client doesn't need explicit cleanup
	return nil
}