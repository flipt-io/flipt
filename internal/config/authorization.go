package config

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*AuthorizationConfig)(nil)
	_ validator = (*AuthorizationConfig)(nil)
)

// AuthorizationConfig configures Flipts authorization mechanisms
type AuthorizationConfig struct {
	// Required designates whether authorization credentials are validated.
	// If required == true, then authorization is required for all API endpoints.
	Required bool                             `json:"required,omitempty" mapstructure:"required" yaml:"required,omitempty"`
	Backend  AuthorizationBackend             `json:"backend,omitempty" mapstructure:"backend" yaml:"backend,omitempty"`
	Local    *AuthorizationLocalConfig        `json:"local,omitempty" mapstructure:"local,omitempty" yaml:"local,omitempty"`
	Cloud    *AuthorizationSourceCloudConfig  `json:"cloud,omitempty" mapstructure:"cloud,omitempty" yaml:"cloud,omitempty"`
	Bundle   *AuthorizationSourceBundleConfig `json:"bundle,omitempty" mapstructure:"bundle,omitempty" yaml:"bundle,omitempty"`
	Object   *AuthorizationSourceObjectConfig `json:"object,omitempty" mapstructure:"object,omitempty" yaml:"object,omitempty"`
}

func (c *AuthorizationConfig) setDefaults(v *viper.Viper) error {
	auth := map[string]any{"required": false}
	if v.GetBool("authorization.required") {
		switch v.GetString("authorization.backend") {
		case string(AuthorizationBackendCloud):
			auth["cloud"] = map[string]any{
				"poll_interval": "5m",
			}
		default:
			auth["backend"] = AuthorizationBackendLocal
			if v.GetString("authorization.local.data.path") == "" {
				auth["local"] = map[string]any{
					"policy": map[string]any{
						"poll_interval": "5m",
					},
				}
			} else {
				auth["local"] = map[string]any{
					"policy": map[string]any{
						"poll_interval": "5m",
					},
					"data": map[string]any{
						"poll_interval": "30s",
					},
				}
			}
		}

		v.SetDefault("authorization", auth)
	}

	return nil
}

func (c *AuthorizationConfig) validate() error {
	if c.Required {
		switch c.Backend {
		case AuthorizationBackendLocal:
			if c.Local == nil {
				return errors.New("authorization: local backend must be configured")
			}

			if err := c.Local.validate(); err != nil {
				return fmt.Errorf("authorization: local: %w", err)
			}

		case AuthorizationBackendCloud:
			if c.Cloud == nil {
				return errors.New("authorization: cloud backend must be configured")
			}

			if err := c.Cloud.validate(); err != nil {
				return fmt.Errorf("authorization: cloud: %w", err)
			}

		case AuthorizationBackendBundle:
			if c.Bundle == nil {
				return errors.New("authorization: bundle backend must be configured")
			}

			if err := c.Bundle.validate(); err != nil {
				return fmt.Errorf("authorization: bundle: %w", err)
			}

		case AuthorizationBackendObject:
			if c.Object == nil {
				return errors.New("authorization: object backend must be configured")
			}

			if err := c.Object.validate(); err != nil {
				return fmt.Errorf("authorization: object: %w", err)
			}
		}
	}

	return nil
}

// AuthorizationBackend configures the data source backend for
// additional data supplied to the policy evaluation engine
type AuthorizationBackend string

const (
	AuthorizationBackendLocal  = AuthorizationBackend("local")
	AuthorizationBackendCloud  = AuthorizationBackend("cloud")
	AuthorizationBackendObject = AuthorizationBackend("object")
	AuthorizationBackendBundle = AuthorizationBackend("bundle")
)

type ObjectAuthorizationBackendType string

const (
	S3ObjectAuthorizationBackendType = ObjectAuthorizationBackendType("s3")
	// AZObjectSubAuthorizationBackendType = ObjectSubAuthorizationBackendType("azblob")
	// GSObjectSubAuthorizationBackendType = ObjectSubAuthorizationBackendType("googlecloud")
)

// AuthorizationLocalConfig configures the local backend source
// for the authorization evaluation engines policies and data
type AuthorizationLocalConfig struct {
	Policy *AuthorizationSourceLocalConfig `json:"policy,omitempty" mapstructure:"policy,omitempty" yaml:"policy,omitempty"`
	Data   *AuthorizationSourceLocalConfig `json:"data,omitempty" mapstructure:"data,omitempty" yaml:"data,omitempty"`
}

func (c *AuthorizationLocalConfig) validate() error {
	if c.Policy == nil {
		return errors.New("policy source must be configured")
	}

	if err := c.Policy.validate(); err != nil {
		return fmt.Errorf("policy: %w", err)
	}

	if err := c.Data.validate(); err != nil {
		return fmt.Errorf("data: %w", err)
	}

	return nil
}

type AuthorizationSourceLocalConfig struct {
	Path         string        `json:"path,omitempty" mapstructure:"path" yaml:"path,omitempty"`
	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
}

func (a *AuthorizationSourceLocalConfig) validate() error {
	if a == nil {
		return nil
	}

	if a.Path == "" {
		return errors.New("path must be non-empty string")
	}

	if a.PollInterval <= 0 {
		return errors.New("poll_interval must be greater than zero")
	}

	return nil
}

type AuthorizationSourceCloudConfig struct {
	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
}

func (a *AuthorizationSourceCloudConfig) validate() error {
	if a != nil && a.PollInterval <= 0 {
		return errors.New("poll_interval must be greater than zero")
	}

	return nil
}

type AuthorizationSourceBundleConfig struct {
	Configuration string `json:"configuration,omitempty" mapstructure:"configuration" yaml:"configuration,omitempty"`
}

func (a *AuthorizationSourceBundleConfig) validate() error {
	if a == nil {
		return nil
	}

	if a.Configuration == "" {
		return errors.New("configuration must be non-empty string")
	}

	return nil
}

// String returns the configuration as a string in the format expected by the OPA engine
func (a *AuthorizationSourceBundleConfig) String() string {
	return a.Configuration
}

// AuthorizationSourceObjectConfig contains authz bundle configuration from object storage.
type AuthorizationSourceObjectConfig struct {
	Type ObjectAuthorizationBackendType `json:"type,omitempty" mapstructure:"type" yaml:"type,omitempty"`
	S3   *S3AuthorizationSource         `json:"s3,omitempty" mapstructure:"s3,omitempty" yaml:"s3,omitempty"`
	// AZBlob *AZBlobAuthorizationSource        `json:"azblob,omitempty" mapstructure:"azblob,omitempty" yaml:"azblob,omitempty"`
	// GS     *GSAuthorizationSource            `json:"googlecloud,omitempty" mapstructure:"googlecloud,omitempty" yaml:"googlecloud,omitempty"`
}

// validate is only called if authorization.backend == "object"
func (a *AuthorizationSourceObjectConfig) validate() error {
	if a == nil {
		return nil
	}

	switch a.Type {
	case S3ObjectAuthorizationBackendType:
		if a.S3 == nil || a.S3.Bucket == "" {
			return errors.New("s3 bucket must be specified")
		}

	default:
		return errors.New("object storage type must be specified")
	}
	return nil
}

// String returns the configuration as a string in the format expected by the OPA engine
func (a *AuthorizationSourceObjectConfig) String() string {
	var (
		url  strings.Builder
		tmpl string
	)

	switch a.Type { //nolint
	case S3ObjectAuthorizationBackendType:
		if a.S3.Endpoint != "" {
			url.WriteString(a.S3.Endpoint)
		} else {
			url.WriteString("https://s3")
			if a.S3.Region != "" {
				url.WriteString(".")
				url.WriteString(a.S3.Region)
			}
			url.WriteString(".amazonaws.com")
		}

		tmpl = fmt.Sprintf(`
services:
  s3:
    url: %s
    credentials:
      s3_signing:
        environment_credentials: {}

bundles:
  flipt:
    service: s3
    resource: %s
`, url.String(), path.Join(a.S3.Bucket, a.S3.Prefix, "bundle.tar.gz"))
	}

	return tmpl
}

// S3AuthorizationSource contains configuration for referencing a s3 bucket
type S3AuthorizationSource struct {
	Endpoint string `json:"-" mapstructure:"endpoint" yaml:"endpoint,omitempty"`
	Bucket   string `json:"bucket,omitempty" mapstructure:"bucket" yaml:"bucket,omitempty"`
	Prefix   string `json:"prefix,omitempty" mapstructure:"prefix" yaml:"prefix,omitempty"`
	Region   string `json:"region,omitempty" mapstructure:"region" yaml:"region,omitempty"`
}

// AZBlobSAuthorizationSource contains configuration for referencing a Azure Blob Storage
// type AZBlobAuthorizationSource struct {
// 	Endpoint     string        `json:"-" mapstructure:"endpoint" yaml:"endpoint,omitempty"`
// 	Container    string        `json:"container,omitempty" mapstructure:"container" yaml:"container,omitempty"`
// }

// GSAuthorizationSource contains configuration for referencing a Google Cloud Storage
// type GSAuthorizationSource struct {
// 	Bucket       string        `json:"-" mapstructure:"bucket" yaml:"bucket,omitempty"`
// 	Prefix       string        `json:"prefix,omitempty" mapstructure:"prefix" yaml:"prefix,omitempty"`
// }
