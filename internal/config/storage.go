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
)

// StorageConfig contains fields which will configure the type of backend in which Flipt will serve
// flag state.
type StorageConfig struct {
	Type  StorageType `json:"type,omitempty" mapstructure:"type"`
	Local Local       `json:"local,omitempty" mapstructure:"local"`
	Git   Git         `json:"git,omitempty" mapstructure:"git"`
}

func (c *StorageConfig) setDefaults(v *viper.Viper) {
	switch v.GetString("storage.type") {
	case string(LocalStorageType):
		v.SetDefault("storage.local.path", ".")
	case string(GitStorageType):
		v.SetDefault("storage.git.ref", "main")
		v.SetDefault("storage.git.poll_interval", "30s")
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
	Authentication Authentication `json:"authentication,omitempty" mapstructure:"authentication"`
}

// Authentication holds structures for various types of auth we support.
// Token auth will take priority over Basic auth if both are provided.
//
// To make things easier, if there are multiple inputs that a particular auth method needs, and
// not all inputs are given but only partially, we will return a validation error.
// (e.g. if username for basic auth is given, and token is also given a validation error will be returned)
type Authentication struct {
	BasicAuth *BasicAuth `json:"basic,omitempty" mapstructure:"basic"`
	TokenAuth *TokenAuth `json:"token,omitempty" mapstructure:"token"`
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
