package secrets

import (
	"context"
	"fmt"
)

// Provider defines the interface for secret storage backends.
type Provider interface {
	// GetSecret retrieves a secret by path.
	GetSecret(ctx context.Context, path string) (*Secret, error)

	// PutSecret stores a secret at the given path.
	PutSecret(ctx context.Context, path string, secret *Secret) error

	// DeleteSecret removes a secret at the given path.
	DeleteSecret(ctx context.Context, path string) error

	// ListSecrets returns all secret paths matching the prefix.
	ListSecrets(ctx context.Context, pathPrefix string) ([]string, error)
}

// Secret represents a stored secret with its data and metadata.
type Secret struct {
	// Path is the secret's location in the provider.
	Path string

	// Data contains the actual secret values as key-value pairs.
	Data map[string][]byte

	// Metadata contains additional information about the secret.
	Metadata map[string]string

	// Version identifies this specific version of the secret.
	Version string
}

// Reference identifies a specific value within a secret.
type Reference struct {
	// Provider is the name of the registered provider to use.
	Provider string `json:"provider" yaml:"provider"`

	// Path is the secret path within the provider.
	Path string `json:"path" yaml:"path"`

	// Key is the specific data key within the secret.
	Key string `json:"key" yaml:"key"`
}

// Validate ensures the reference is properly configured.
func (r Reference) Validate() error {
	if r.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if r.Path == "" {
		return fmt.Errorf("path is required")
	}
	if r.Key == "" {
		return fmt.Errorf("key is required")
	}
	return nil
}

// GetValue retrieves a specific value from the secret data.
func (s *Secret) GetValue(key string) ([]byte, bool) {
	value, ok := s.Data[key]
	return value, ok
}

// SetValue sets a specific value in the secret data.
func (s *Secret) SetValue(key string, value []byte) {
	if s.Data == nil {
		s.Data = make(map[string][]byte)
	}
	s.Data[key] = value
}

// GetMetadata retrieves a metadata value by key.
func (s *Secret) GetMetadata(key string) (string, bool) {
	value, ok := s.Metadata[key]
	return value, ok
}

// SetMetadata sets a metadata value.
func (s *Secret) SetMetadata(key, value string) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}
	s.Metadata[key] = value
}
