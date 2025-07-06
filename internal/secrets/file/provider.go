package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/secrets"
	"go.uber.org/zap"
)

func init() {
	// Register file provider factory
	secrets.RegisterProviderFactory("file", func(cfg *config.Config, logger *zap.Logger) (secrets.Provider, error) {
		if cfg.Secrets.Providers.File == nil {
			return nil, fmt.Errorf("file provider configuration not found")
		}
		return NewProvider(cfg.Secrets.Providers.File.BasePath, logger)
	})
}

// Provider implements secrets.Provider for file-based storage.
type Provider struct {
	basePath string
	logger   *zap.Logger
}

// NewProvider creates a new file-based secret provider.
func NewProvider(basePath string, logger *zap.Logger) (*Provider, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0700); err != nil {
		return nil, fmt.Errorf("creating base path: %w", err)
	}

	return &Provider{
		basePath: basePath,
		logger:   logger,
	}, nil
}

// GetSecret retrieves a secret from the file system.
func (p *Provider) GetSecret(ctx context.Context, path string) (*secrets.Secret, error) {
	fullPath := p.secretPath(path)

	p.logger.Debug("reading secret from file",
		zap.String("path", path),
		zap.String("file", fullPath))

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("secret not found at path %q", path)
		}
		return nil, fmt.Errorf("reading secret file: %w", err)
	}

	var fileData secretFile
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("parsing secret file: %w", err)
	}

	// Convert string data to bytes
	secretData := make(map[string][]byte)
	for k, v := range fileData.Data {
		secretData[k] = []byte(v)
	}

	return &secrets.Secret{
		Path:     path,
		Data:     secretData,
		Metadata: fileData.Metadata,
		Version:  fileData.Version,
	}, nil
}

// PutSecret stores a secret in the file system.
func (p *Provider) PutSecret(ctx context.Context, path string, secret *secrets.Secret) error {
	fullPath := p.secretPath(path)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Convert bytes to strings for JSON serialization
	stringData := make(map[string]string)
	for k, v := range secret.Data {
		stringData[k] = string(v)
	}

	fileData := secretFile{
		Data:     stringData,
		Metadata: secret.Metadata,
		Version:  secret.Version,
	}

	data, err := json.MarshalIndent(fileData, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding secret: %w", err)
	}

	p.logger.Debug("writing secret to file",
		zap.String("path", path),
		zap.String("file", fullPath))

	// Write with restricted permissions
	if err := os.WriteFile(fullPath, data, 0600); err != nil {
		return fmt.Errorf("writing secret file: %w", err)
	}

	return nil
}

// DeleteSecret removes a secret from the file system.
func (p *Provider) DeleteSecret(ctx context.Context, path string) error {
	fullPath := p.secretPath(path)

	p.logger.Debug("deleting secret file",
		zap.String("path", path),
		zap.String("file", fullPath))

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("secret not found at path %q", path)
		}
		return fmt.Errorf("deleting secret file: %w", err)
	}

	// Clean up empty directories
	dir := filepath.Dir(fullPath)
	for dir != p.basePath {
		if err := os.Remove(dir); err != nil {
			// Directory not empty, stop cleanup
			break
		}
		dir = filepath.Dir(dir)
	}

	return nil //nolint:nilerr // cleanup errors are not critical
}

// ListSecrets returns all secret paths matching the prefix.
func (p *Provider) ListSecrets(ctx context.Context, pathPrefix string) ([]string, error) {
	var paths []string

	searchPath := filepath.Join(p.basePath, pathPrefix)

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil //nolint:nilerr // skip inaccessible paths in walk
		}

		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Convert file path back to secret path
		relPath, err := filepath.Rel(p.basePath, path)
		if err != nil {
			return nil //nolint:nilerr // skip invalid paths in walk
		}

		// Remove .json extension
		secretPath := strings.TrimSuffix(relPath, ".json")
		paths = append(paths, secretPath)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("listing secrets: %w", err)
	}

	p.logger.Debug("listed secrets",
		zap.String("prefix", pathPrefix),
		zap.Int("count", len(paths)))

	return paths, nil
}

// secretPath converts a logical path to a file system path.
func (p *Provider) secretPath(path string) string {
	// Sanitize path to prevent directory traversal
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")

	return filepath.Join(p.basePath, path+".json")
}

// secretFile represents the JSON structure stored on disk.
type secretFile struct {
	Data     map[string]string `json:"data"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Version  string            `json:"version,omitempty"`
}
