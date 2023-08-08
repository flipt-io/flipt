package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

type StorageType string

const (
	DatabaseStorageType = StorageType("database")
	LocalStorageType    = StorageType("local")
	GitStorageType      = StorageType("git")
	ObjectStorageType   = StorageType("object")
)

type ObjectSubStorageType string

const (
	S3ObjectSubStorageType = ObjectSubStorageType("s3")
)

// StorageConfig contains fields which will configure the type of backend in which Flipt will serve
// flag state.
type StorageConfig struct {
	Type   StorageType `json:"type,omitempty" mapstructure:"type"`
	Local  *Local      `json:"local,omitempty" mapstructure:"local,omitempty"`
	Git    *Git        `json:"git,omitempty" mapstructure:"git,omitempty"`
	Object *Object     `json:"object,omitempty" mapstructure:"object,omitempty"`
}

func (c *StorageConfig) setDefaults(v *viper.Viper) {
	switch v.GetString("storage.type") {
	case string(LocalStorageType):
		v.SetDefault("storage.local.path", ".")
	case string(GitStorageType):
		v.SetDefault("storage.git.ref", "main")
		v.SetDefault("storage.git.poll_interval", "30s")
	case string(ObjectStorageType):
		// keep this as a case statement in anticipation of
		// more object types in the future
		// nolint:gocritic
		switch v.GetString("storage.object.type") {
		case string(S3ObjectSubStorageType):
			v.SetDefault("storage.object.s3.poll_interval", "1m")
		}
	default:
		v.SetDefault("storage.type", "database")
	}
}

func (c *StorageConfig) validate() error {
	if c.Type == GitStorageType {
		if c.Git.Ref == "" {
			return errors.New("git ref must be specified")
		}
		if c.Git.Repository == "" {
			return errors.New("git repository must be specified")
		}

		if err := c.Git.Authentication.validate(); err != nil {
			return err
		}
	}

	if c.Type == LocalStorageType {
		if c.Local.Path == "" {
			return errors.New("local path must be specified")
		}
	}

	if c.Type == ObjectStorageType {
		if c.Object == nil {
			return errors.New("object storage type must be specified")
		}
		if err := c.Object.validate(); err != nil {
			return err
		}
	}

	return nil
}

// Local contains configuration for referencing a local filesystem.
type Local struct {
	Path string `json:"path,omitempty" mapstructure:"path"`
}

// Git contains configuration for referencing a git repository.
type Git struct {
	Repository     string         `json:"repository,omitempty" mapstructure:"repository"`
	Ref            string         `json:"ref,omitempty" mapstructure:"ref"`
	PollInterval   time.Duration  `json:"pollInterval,omitempty" mapstructure:"poll_interval"`
	Authentication Authentication `json:"authentication,omitempty" mapstructure:"authentication,omitempty"`
}

// Object contains configuration of readonly object storage.
type Object struct {
	Type ObjectSubStorageType `json:"type,omitempty" mapstructure:"type"`
	S3   *S3                  `json:"s3,omitempty" mapstructure:"s3,omitempty"`
}

// validate is only called if storage.type == "object"
func (o *Object) validate() error {
	switch o.Type {
	case S3ObjectSubStorageType:
		if o.S3 == nil || o.S3.Bucket == "" {
			return errors.New("s3 bucket must be specified")
		}
	default:
		return errors.New("object storage type must be specified")
	}
	return nil
}

// S3 contains configuration for referencing a s3 bucket
type S3 struct {
	Endpoint     string        `json:"endpoint,omitempty" mapstructure:"endpoint"`
	Bucket       string        `json:"bucket,omitempty" mapstructure:"bucket"`
	Prefix       string        `json:"prefix,omitempty" mapstructure:"prefix"`
	Region       string        `json:"region,omitempty" mapstructure:"region"`
	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval"`
}

// Authentication holds structures for various types of auth we support.
// Token auth will take priority over Basic auth if both are provided.
//
// To make things easier, if there are multiple inputs that a particular auth method needs, and
// not all inputs are given but only partially, we will return a validation error.
// (e.g. if username for basic auth is given, and token is also given a validation error will be returned)
type Authentication struct {
	BasicAuth *BasicAuth `json:"basic,omitempty" mapstructure:"basic,omitempty"`
	TokenAuth *TokenAuth `json:"token,omitempty" mapstructure:"token,omitempty"`
}

func (a *Authentication) validate() error {
	if a.BasicAuth != nil {
		if err := a.BasicAuth.validate(); err != nil {
			return err
		}
	}
	if a.TokenAuth != nil {
		if err := a.TokenAuth.validate(); err != nil {
			return err
		}
	}

	return nil
}

// BasicAuth has configuration for authenticating with private git repositories
// with basic auth.
type BasicAuth struct {
	Username string `json:"username,omitempty" mapstructure:"username"`
	Password string `json:"password,omitempty" mapstructure:"password"`
}

func (b BasicAuth) validate() error {
	if (b.Username != "" && b.Password == "") || (b.Username == "" && b.Password != "") {
		return errors.New("both username and password need to be provided for basic auth")
	}

	return nil
}

// TokenAuth has configuration for authenticating with private git repositories
// with token auth.
type TokenAuth struct {
	AccessToken string `json:"accessToken,omitempty" mapstructure:"access_token"`
}

func (t TokenAuth) validate() error { return nil }
