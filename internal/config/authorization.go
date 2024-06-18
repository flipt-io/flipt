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
	Custom   *AuthorizationSourceCustomConfig `json:"custom,omitempty" mapstructure:"custom,omitempty" yaml:"custom,omitempty"`
	Object   *AuthorizationSourceObjectConfig `json:"object,omitempty" mapstructure:"object,omitempty" yaml:"object,omitempty"`
}

func (c *AuthorizationConfig) setDefaults(v *viper.Viper) error {
	auth := map[string]any{"required": false}
	if v.GetBool("authorization.required") {
		switch v.GetString("authorization.backend") {
		case string(AuthorizationBackendObject):
			auth["object"] = map[string]any{
				"type": S3ObjectSubAuthorizationBackendType,
				"s3": map[string]any{
					"poll_interval": "60s",
				},
			}
		default:
			auth["backend"] = AuthorizationBackendLocal
			auth["local.policy"] = map[string]any{
				"poll_interval": "5m",
			}

			if v.GetString("authorization.local.data.path") != "" {
				auth["local.data"] = map[string]any{
					"poll_interval": "30s",
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
		case AuthorizationBackendCustom:
			if c.Custom == nil {
				return errors.New("authorization: custom backend must be configured")
			}

			if err := c.Custom.validate(); err != nil {
				return fmt.Errorf("authorization: custom: %w", err)
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
	AuthorizationBackendObject = AuthorizationBackend("object")
	AuthorizationBackendCustom = AuthorizationBackend("custom")
)

type ObjectSubAuthorizationBackendType string

const (
	S3ObjectSubAuthorizationBackendType = ObjectSubAuthorizationBackendType("s3")
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
	Path         string        `json:"path,omitempty" mapstructure:"path" yaml:"backend,path"`
	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
}

func (a *AuthorizationSourceLocalConfig) validate() error {
	if a == nil {
		return nil
	}

	if a.Path == "" {
		return errors.New("path must be non-empty string")
	}

	if a.Path != "" && a.PollInterval <= 0 {
		return errors.New("poll_interval must be non-zero")
	}

	return nil
}

type AuthorizationSourceCustomConfig struct {
	Configuration string `json:"configuration,omitempty" mapstructure:"configuration" yaml:"configuration,omitempty"`
}

func (a *AuthorizationSourceCustomConfig) validate() error {
	if a == nil {
		return nil
	}

	if a.Configuration == "" {
		return errors.New("configuration must be non-empty string")
	}

	return nil
}

// String returns the configuration as a string in the format expected by the OPA engine
func (a *AuthorizationSourceCustomConfig) String() string {
	return a.Configuration
}

// AuthorizationSourceObjectConfig contains authz bundle configuration from object storage.
type AuthorizationSourceObjectConfig struct {
	Type ObjectSubAuthorizationBackendType `json:"type,omitempty" mapstructure:"type" yaml:"type,omitempty"`
	S3   *S3AuthorizationSource            `json:"s3,omitempty" mapstructure:"s3,omitempty" yaml:"s3,omitempty"`
	// AZBlob *AZBlobAuthorizationSource        `json:"azblob,omitempty" mapstructure:"azblob,omitempty" yaml:"azblob,omitempty"`
	// GS     *GSAuthorizationSource            `json:"googlecloud,omitempty" mapstructure:"googlecloud,omitempty" yaml:"googlecloud,omitempty"`
}

// validate is only called if authorization.backend == "object"
func (a *AuthorizationSourceObjectConfig) validate() error {
	if a == nil {
		return nil
	}

	switch a.Type {
	case S3ObjectSubAuthorizationBackendType:
		if a.S3 == nil || a.S3.Bucket == "" {
			return errors.New("s3 bucket must be specified")
		}

		if a.S3.PollInterval < 60*time.Second {
			return errors.New("s3 poll_interval must be at least 60s")
		}

	// case AZObjectSubAuthorizationBackendType:
	// 	if o.AZBlob == nil || o.AZBlob.Container == "" {
	// 		return errors.New("azblob container must be specified")
	// 	}
	// case GSObjectSubAuthorizationBackendType:
	// 	if o.GS == nil || o.GS.Bucket == "" {
	// 		return errors.New("googlecloud bucket must be specified")
	// 	}
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

	switch a.Type {
	case S3ObjectSubAuthorizationBackendType:
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
    polling:
      min_delay_seconds: 30
      max_delay_seconds: %d
`, url.String(), path.Join(a.S3.Bucket, a.S3.Prefix, "bundle.tar.gz"), int(a.S3.PollInterval.Seconds()))
	}

	return tmpl
}

// S3AuthorizationSource contains configuration for referencing a s3 bucket
type S3AuthorizationSource struct {
	Endpoint     string        `json:"-" mapstructure:"endpoint" yaml:"endpoint,omitempty"`
	Bucket       string        `json:"bucket,omitempty" mapstructure:"bucket" yaml:"bucket,omitempty"`
	Prefix       string        `json:"prefix,omitempty" mapstructure:"prefix" yaml:"prefix,omitempty"`
	Region       string        `json:"region,omitempty" mapstructure:"region" yaml:"region,omitempty"`
	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
}

// AZBlobSAuthorizationSource contains configuration for referencing a Azure Blob Storage
// type AZBlobAuthorizationSource struct {
// 	Endpoint     string        `json:"-" mapstructure:"endpoint" yaml:"endpoint,omitempty"`
// 	Container    string        `json:"container,omitempty" mapstructure:"container" yaml:"container,omitempty"`
// 	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
// }

// GSAuthorizationSource contains configuration for referencing a Google Cloud Storage
// type GSAuthorizationSource struct {
// 	Bucket       string        `json:"-" mapstructure:"bucket" yaml:"bucket,omitempty"`
// 	Prefix       string        `json:"prefix,omitempty" mapstructure:"prefix" yaml:"prefix,omitempty"`
// 	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
// }
